/*
   Copyright (c) 2016, Percona LLC and/or its affiliates. All rights reserved.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package report

import (
	"sort"
	"time"

	"github.com/percona/go-mysql/event"

	pc "github.com/percona/pmm-agent/agents/builtin/mongo/proto/config"
	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/qan"
)

// slowlog|perf schema --> Result --> qan.Report --> data.Spooler

// Data for an interval from slow log or performance schema (pfs) parser,
// passed to MakeReport() which transforms into a qan.Report{}.
type Result struct {
	Global     *event.Class   // metrics for all data
	Class      []*event.Class // per-class metrics
	RateLimit  uint           // Percona Server rate limit
	RunTime    float64        // seconds parsing data, hopefully < interval
	StopOffset int64          // slow log offset where parsing stopped, should be <= end offset
	Error      string         `json:",omitempty"`
}

type ByQueryTime []*event.Class

func (a ByQueryTime) Len() int      { return len(a) }
func (a ByQueryTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByQueryTime) Less(i, j int) bool {
	// todo: will panic if struct is incorrect
	// descending order
	return a[i].Metrics.TimeMetrics["Query_time"].Sum > a[j].Metrics.TimeMetrics["Query_time"].Sum
}

func MakeReport(config pc.QAN, startTime, endTime time.Time, result *Result) *qan.Report {
	// Sort classes by Query_time_sum, descending.
	sort.Sort(ByQueryTime(result.Class))

	// Make qan.Report from Result and other metadata (e.g. Interval).
	report := &qan.Report{
		UUID:    config.UUID,
		StartTs: startTime,
		EndTs:   endTime,
		RunTime: result.RunTime,
		Global:  result.Global,
		Class:   result.Class,
	}

	// Return all query classes if there's no limit or number of classes is
	// less than the limit.
	n := len(result.Class)
	if config.ReportLimit == 0 || n <= int(config.ReportLimit) {
		return report // all classes, no LRQ
	}

	// Top queries
	report.Class = result.Class[0:config.ReportLimit]

	// Low-ranking Queries
	lrq := event.NewClass("lrq", "/* low-ranking queries */", false)
	for _, class := range result.Class[config.ReportLimit:n] {
		lrq.AddClass(class)
	}
	report.Class = append(report.Class, lrq)

	return report // top classes, the rest as LRQ
}
