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
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/percona/go-mysql/event"
	pc "github.com/percona/pmm/proto/config"
	"github.com/percona/qan-agent/qan/analyzer/mysql/iter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var outputDir = RootDir() + "/test/qan/"

func TestResult001(t *testing.T) {
	data, err := ioutil.ReadFile(outputDir + "/result001.json")
	require.NoError(t, err)

	result := &Result{}
	err = json.Unmarshal(data, result)
	require.NoError(t, err)

	start := time.Now().Add(-1 * time.Second)
	stop := time.Now()

	interval := &iter.Interval{
		Filename:    "slow.log",
		StartTime:   start,
		StopTime:    stop,
		StartOffset: 0,
		EndOffset:   1000,
	}
	config := pc.QAN{
		UUID:        "1",
		ReportLimit: 10,
	}
	report := MakeReport(config, interval.StartTime, interval.StopTime, interval, result)

	// 1st: 2.9
	assert.Equal(t, "3000000000000003", report.Class[0].Id)
	assert.Equal(t, float64(2.9), report.Class[0].Metrics.TimeMetrics["Query_time"].Sum)
	// 2nd: 2
	assert.Equal(t, "2000000000000002", report.Class[1].Id)
	assert.Equal(t, float64(2), report.Class[1].Metrics.TimeMetrics["Query_time"].Sum)
	// ...
	// 5th: 0.101001
	assert.Equal(t, "5000000000000005", report.Class[4].Id)
	assert.Equal(t, float64(0.101001), report.Class[4].Metrics.TimeMetrics["Query_time"].Sum)

	// Limit=2 results in top 2 queries and the rest in 1 LRQ "query".
	config.ReportLimit = 2
	report = MakeReport(config, interval.StartTime, interval.StopTime, interval, result)
	assert.Equal(t, 3, len(report.Class))

	assert.Equal(t, "3000000000000003", report.Class[0].Id)
	assert.Equal(t, float64(2.9), report.Class[0].Metrics.TimeMetrics["Query_time"].Sum)

	assert.Equal(t, "2000000000000002", report.Class[1].Id)
	assert.Equal(t, float64(2), report.Class[1].Metrics.TimeMetrics["Query_time"].Sum)

	assert.Equal(t, "lrq", report.Class[2].Id)
	assert.Equal(t, 10, int(report.Class[2].TotalQueries))
	assert.Equal(t, float64(1+1+0.101001), report.Class[2].Metrics.TimeMetrics["Query_time"].Sum)
	assert.Equal(t, event.Float64(0.000100), report.Class[2].Metrics.TimeMetrics["Query_time"].Min)
	assert.Equal(t, event.Float64(1.12), report.Class[2].Metrics.TimeMetrics["Query_time"].Max)
	assert.Equal(t, event.Float64((1+1+0.101001)/10), report.Class[2].Metrics.TimeMetrics["Query_time"].Avg)
}
