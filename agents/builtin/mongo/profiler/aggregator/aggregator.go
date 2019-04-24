// pmm-agent
// Copyright (C) 2018 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package aggregator

import (
	"fmt"
	"sync"
	"time"

	"github.com/percona/percona-toolkit/src/go/mongolib/fingerprinter"
	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	mongostats "github.com/percona/percona-toolkit/src/go/mongolib/stats"
	"github.com/percona/pmm/api/qanpb"

	pc "github.com/percona/pmm-agent/agents/builtin/mongo/proto/config"
	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/qan"
	"github.com/percona/pmm-agent/agents/builtin/mongo/status"
)

const (
	DefaultInterval       = 60 // in seconds
	DefaultExampleQueries = true
	ReportChanBuffer      = 1000
)

// New returns configured *Aggregator
func New(timeStart time.Time, config pc.QAN) *Aggregator {
	defaultExampleQueries := DefaultExampleQueries
	// verify config
	if config.Interval == 0 {
		config.Interval = DefaultInterval
		config.ExampleQueries = &defaultExampleQueries
	}

	aggregator := &Aggregator{
		config: config,
	}

	// create duration from interval
	aggregator.d = time.Duration(config.Interval) * time.Second

	// create mongolib stats
	fp := fingerprinter.NewFingerprinter(fingerprinter.DEFAULT_KEY_FILTERS)
	aggregator.mongostats = mongostats.New(fp)

	// create new interval
	aggregator.newInterval(timeStart)

	return aggregator
}

// Aggregator aggregates system.profile document
type Aggregator struct {
	// dependencies
	config pc.QAN

	// status
	status *status.Status
	stats  *stats

	// provides
	reportChan chan *qan.Report

	// interval
	timeStart  time.Time
	timeEnd    time.Time
	d          time.Duration
	t          *time.Timer
	mongostats *mongostats.Stats

	// state
	sync.RWMutex                 // Lock() to protect internal consistency of the service
	running      bool            // Is this service running?
	doneChan     chan struct{}   // close(doneChan) to notify goroutines that they should shutdown
	wg           *sync.WaitGroup // Wait() for goroutines to stop after being notified they should shutdown
}

// Add aggregates new system.profile document
func (a *Aggregator) Add(doc proto.SystemProfile) error {
	a.Lock()
	defer a.Unlock()
	if !a.running {
		return fmt.Errorf("aggregator is not running")
	}

	ts := doc.Ts.UTC()

	// skip old metrics
	if ts.Before(a.timeStart) {
		a.stats.DocsSkippedOld.Add(1)
		return nil
	}

	// if new doc is outside of interval then finish old interval and flush it
	if !ts.Before(a.timeEnd) {
		a.flush(ts)
	}

	// we had some activity so reset timer
	a.t.Reset(a.d)

	// add new doc to stats
	a.stats.DocsIn.Add(1)
	return a.mongostats.Add(doc)
}

func (a *Aggregator) Start() <-chan *qan.Report {
	a.Lock()
	defer a.Unlock()
	if a.running {
		return a.reportChan
	}

	// create new channels over which we will communicate to...
	// ... outside world by sending collected docs
	a.reportChan = make(chan *qan.Report, ReportChanBuffer)
	// ... inside goroutine to close it
	a.doneChan = make(chan struct{})

	// set status
	a.stats = &stats{}
	a.status = status.New(a.stats)

	// timeout after not receiving data for interval time
	a.t = time.NewTimer(a.d)

	// start a goroutine and Add() it to WaitGroup
	// so we could later Wait() for it to finish
	a.wg = &sync.WaitGroup{}
	a.wg.Add(1)
	go start(
		a.wg,
		a,
		a.doneChan,
		a.stats,
	)

	a.running = true
	return a.reportChan
}

func (a *Aggregator) Stop() {
	a.Lock()
	defer a.Unlock()
	if !a.running {
		return
	}
	a.running = false

	// notify goroutine to close
	close(a.doneChan)

	// wait for goroutines to exit
	a.wg.Wait()

	// close reportChan
	close(a.reportChan)
}

func (a *Aggregator) Status() map[string]string {
	a.RLock()
	defer a.RUnlock()
	if !a.running {
		return nil
	}

	return a.status.Map()
}

func start(
	wg *sync.WaitGroup,
	aggregator *Aggregator,
	doneChan <-chan struct{},
	stats *stats,
) {
	// signal WaitGroup when goroutine finished
	defer wg.Done()

	// update stats
	stats.IntervalStart.Set(aggregator.TimeStart().Format("2006-01-02 15:04:05"))
	stats.IntervalEnd.Set(aggregator.TimeEnd().Format("2006-01-02 15:04:05"))
	for {
		select {
		case <-aggregator.t.C:
			// When Tail()ing system.profile collection you don't know if sample
			// is last sample in the collection until you get sample with higher timestamp than interval.
			// For this, in cases where we generate only few test queries,
			// but still expect them to show after interval expires, we need to implement timeout.
			// This introduces another issue, that in case something goes wrong, and we get metrics for old interval too late, they will be skipped.
			// A proper solution would be to allow fixing old samples, but API and qan-agent doesn't allow this, yet.
			aggregator.Flush()
		case <-doneChan:
			// Check if we should shutdown.
			return
		}
	}
}

func (a *Aggregator) Flush() {
	a.Lock()
	defer a.Unlock()
	a.flush(time.Now())
}

func (a *Aggregator) flush(ts time.Time) {
	r := a.interval(ts)
	if r != nil {
		a.reportChan <- r
		a.stats.ReportsOut.Add(1)
	}
}

// interval sets interval if necessary and returns *qan.Report for old interval if not empty
func (a *Aggregator) interval(ts time.Time) *qan.Report {
	// create new interval
	defer a.newInterval(ts)

	// let's check if we have anything to send for current interval
	if len(a.mongostats.Queries()) == 0 {
		// if there are no queries then we don't create report #PMM-927
		return nil
	}

	// create result
	result := a.createResult()

	// translate result into report and return it
	return qan.MakeReport(a.config, a.timeStart, a.timeEnd, result)
}

// TimeStart returns start time for current interval
func (a *Aggregator) TimeStart() time.Time {
	return a.timeStart
}

// TimeEnd returns end time for current interval
func (a *Aggregator) TimeEnd() time.Time {
	return a.timeEnd
}

func (a *Aggregator) newInterval(ts time.Time) {
	// reset stats
	a.mongostats.Reset()

	// truncate to the duration e.g 12:15:35 with 1 minute duration it will be 12:15:00
	a.timeStart = ts.UTC().Truncate(a.d)
	// create ending time by adding interval
	a.timeEnd = a.timeStart.Add(a.d)
}

func (a *Aggregator) createResult() *qan.Result {
	queries := a.mongostats.Queries()
	queryStats := queries.CalcQueriesStats(int64(a.config.Interval))
	buckets := []*qanpb.MetricsBucket{}

	for _, queryInfo := range queryStats {
		bucket := &qanpb.MetricsBucket{}
		// TODO: Add more metrics... (See: https://jira.percona.com/browse/PMM-3880)

		bucket.Queryid = queryInfo.ID
		bucket.Fingerprint = queryInfo.Fingerprint

		bucket.NumQueries = float32(queryInfo.Count)

		bucket.MDocsReturnedCnt = float32(queryInfo.Count)
		bucket.MDocsReturnedMax = float32(queryInfo.Returned.Max)
		bucket.MDocsReturnedMin = float32(queryInfo.Returned.Min)
		bucket.MDocsReturnedP99 = float32(queryInfo.Returned.Pct) // TODO: Replace to P99
		bucket.MDocsReturnedSum = float32(queryInfo.Returned.Total)

		bucket.MDocsScannedCnt = float32(queryInfo.Count)
		bucket.MDocsScannedMax = float32(queryInfo.Scanned.Max)
		bucket.MDocsScannedMin = float32(queryInfo.Scanned.Min)
		bucket.MDocsScannedP99 = float32(queryInfo.Scanned.Pct) // TODO: Replace to P99
		bucket.MDocsScannedSum = float32(queryInfo.Scanned.Total)

		bucket.MResponseLengthCnt = float32(queryInfo.Count)
		bucket.MResponseLengthMax = float32(queryInfo.ResponseLength.Max)
		bucket.MResponseLengthMin = float32(queryInfo.ResponseLength.Min)
		bucket.MResponseLengthP99 = float32(queryInfo.ResponseLength.Pct) // TODO: Replace to P99
		bucket.MResponseLengthSum = float32(queryInfo.ResponseLength.Total)

		buckets = append(buckets, bucket)
	}

	return &qan.Result{
		Buckets: buckets,
	}

}
