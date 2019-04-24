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

package report

import (
	"sort"
	"time"

	"github.com/percona/pmm/api/qanpb"

	"github.com/percona/pmm-agent/agents/builtin/mongo/config"
)

type Report struct {
	UUID    string                 // UUID of MySQL instance
	StartTs time.Time              // Start time of interval, UTC
	EndTs   time.Time              // Stop time of interval, UTC
	RunTime float64                // Time parsing data, seconds
	Buckets []*qanpb.MetricsBucket // per-class metrics
}

// slowlog|perf schema --> Result --> qan.Report --> data.Spooler

// Data for an interval from slow log or performance schema (pfs) parser,
// passed to MakeReport() which transforms into a qan.Report{}.
type Result struct {
	Buckets    []*qanpb.MetricsBucket
	RateLimit  uint    // Percona Server rate limit
	RunTime    float64 // seconds parsing data, hopefully < interval
	StopOffset int64   // slow log offset where parsing stopped, should be <= end offset
	Error      string  `json:",omitempty"`
}

type ByQueryTime []*qanpb.MetricsBucket

func (a ByQueryTime) Len() int      { return len(a) }
func (a ByQueryTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByQueryTime) Less(i, j int) bool {
	// todo: will panic if struct is incorrect
	// descending order
	return a[i].MQueryTimeSum > a[j].MQueryTimeSum
}

func MakeReport(config config.QAN, startTime, endTime time.Time, result *Result) *Report {
	// Sort classes by Query_time_sum, descending.
	sort.Sort(ByQueryTime(result.Buckets))

	// Make qan.Report from Result and other metadata (e.g. Interval).
	report := &Report{
		UUID:    config.UUID,
		StartTs: startTime,
		EndTs:   endTime,
		RunTime: result.RunTime,
		Buckets: result.Buckets,
	}

	return report
}
