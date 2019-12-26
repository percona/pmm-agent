// pmm-agent
// Copyright 2019 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aggregator

import (
	"testing"
	"time"

	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/api/inventorypb"
	"github.com/stretchr/testify/require"

	"github.com/percona/pmm-agent/agents/mongodb/internal/report"

	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAggregator(t *testing.T) {
	// we need at least one test per package to correctly calculate coverage
	t.Run("Add", func(t *testing.T) {
		t.Run("error if aggregator is not running", func(t *testing.T) {
			a := New(time.Now(), "test-agent", logrus.WithField("component", "test"))
			err := a.Add(proto.SystemProfile{})
			assert.EqualError(t, err, "aggregator is not running")
		})
	})

	t.Run("createResult", func(t *testing.T) {
		agentID := "test-agent"
		startPeriod := time.Now()
		aggregator := New(startPeriod, agentID, logrus.WithField("component", "test"))
		aggregator.Start()
		defer aggregator.Stop()
		err := aggregator.Add(proto.SystemProfile{
			NscannedObjects: 2,
			Nreturned:       3,
			Ns:              "collection.people",
			Op:              "insert",
		})
		require.NoError(t, err)

		result := aggregator.createResult()

		require.Equal(t, 1, len(result.Buckets))
		assert.Equal(t, report.Result{
			Buckets: []*agentpb.MetricsBucket{
				{
					Common: &agentpb.MetricsBucket_Common{
						Queryid:             result.Buckets[0].Common.Queryid,
						Fingerprint:         "INSERT people",
						Database:            "collection",
						Tables:              []string{"people"},
						AgentId:             agentID,
						AgentType:           inventorypb.AgentType_QAN_MONGODB_PROFILER_AGENT,
						PeriodStartUnixSecs: uint32(startPeriod.Truncate(DefaultInterval).Unix()),
						PeriodLengthSecs:    60,
						Example:             `{"ns":"collection.people","op":"insert"}`,
						ExampleFormat:       agentpb.ExampleFormat_EXAMPLE,
						ExampleType:         agentpb.ExampleType_RANDOM,
						NumQueries:          1,
						MQueryTimeCnt:       1,
					},
					Mongodb: &agentpb.MetricsBucket_MongoDB{
						MDocsReturnedCnt:   1,
						MDocsReturnedSum:   3,
						MDocsReturnedMin:   3,
						MDocsReturnedMax:   3,
						MDocsReturnedP99:   3,
						MResponseLengthCnt: 1,
						MDocsScannedCnt:    1,
						MDocsScannedSum:    2,
						MDocsScannedMin:    2,
						MDocsScannedMax:    2,
						MDocsScannedP99:    2,
					},
				},
			},
		}, *result)
	})
}
