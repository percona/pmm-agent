package aggregator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/percona/go-mysql/event"
	"github.com/percona/percona-toolkit/src/go/mongolib/fingerprinter"
	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	mongostats "github.com/percona/percona-toolkit/src/go/mongolib/stats"

	pc "github.com/percona/pmm-agent/agents/builtin/mongo/proto/config"
	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/qan"
	"github.com/percona/pmm-agent/agents/builtin/mongo/report"
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
	return report.MakeReport(a.config, a.timeStart, a.timeEnd, result)
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

// TODO: Here we create mapping from MongoDB metrics to MySQLSlowLog report.
// TODO: This should be refactored.
func (a *Aggregator) createResult() *report.Result {
	queries := a.mongostats.Queries()
	global := event.NewClass("", "", false)
	queryStats := queries.CalcQueriesStats(int64(a.config.Interval))
	classes := []*event.Class{}
	exampleQueries := boolValue(a.config.ExampleQueries)
	for _, queryInfo := range queryStats {
		class := event.NewClass(queryInfo.ID, queryInfo.Fingerprint, exampleQueries)
		if exampleQueries {
			db := ""
			s := strings.SplitN(queryInfo.Namespace, ".", 2)
			if len(s) == 2 {
				db = s[0]
			}

			class.Example = &event.Example{
				QueryTime: queryInfo.QueryTime.Total,
				Db:        db,
				Query:     queryInfo.Query,
			}
		}

		metrics := event.NewMetrics()

		metrics.TimeMetrics["Query_time"] = newEventTimeStatsInMilliseconds(queryInfo.QueryTime)

		// @todo we map below metrics to MySQL equivalents according to PMM-830
		metrics.NumberMetrics["Bytes_sent"] = newEventNumberStats(queryInfo.ResponseLength)
		metrics.NumberMetrics["Rows_sent"] = newEventNumberStats(queryInfo.Returned)
		metrics.NumberMetrics["Rows_examined"] = newEventNumberStats(queryInfo.Scanned)

		class.Metrics = metrics
		class.TotalQueries = uint(queryInfo.Count)
		class.UniqueQueries = 1
		classes = append(classes, class)

		// Add the class to the global metrics.
		global.AddClass(class)
	}

	return &report.Result{
		Global: global,
		Class:  classes,
	}

}

func newEventNumberStats(s mongostats.Statistics) *event.NumberStats {
	return &event.NumberStats{
		Sum: uint64(s.Total),
		Min: event.Uint64(uint64(s.Min)),
		Avg: event.Uint64(uint64(s.Avg)),
		Med: event.Uint64(uint64(s.Median)),
		P95: event.Uint64(uint64(s.Pct95)),
		Max: event.Uint64(uint64(s.Max)),
	}
}

func newEventTimeStatsInMilliseconds(s mongostats.Statistics) *event.TimeStats {
	return &event.TimeStats{
		Sum: s.Total / 1000,
		Min: event.Float64(s.Min / 1000),
		Avg: event.Float64(s.Avg / 1000),
		Med: event.Float64(s.Median / 1000),
		P95: event.Float64(s.Pct95 / 1000),
		Max: event.Float64(s.Max / 1000),
	}
}

// boolValue returns the value of the bool pointer passed in or
// false if the pointer is nil.
func boolValue(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}
