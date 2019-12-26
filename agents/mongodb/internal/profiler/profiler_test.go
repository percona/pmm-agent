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

package profiler

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/api/inventorypb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/percona/pmm-agent/agents/mongodb/internal/profiler/aggregator"
	"github.com/percona/pmm-agent/agents/mongodb/internal/report"
)

func TestProfiler(t *testing.T) {
	defaultInterval := aggregator.DefaultInterval
	aggregator.DefaultInterval = time.Duration(time.Second * 30)
	defer func() { aggregator.DefaultInterval = defaultInterval }()

	url := "mongodb://root:root-password@127.0.0.1:27017"

	sess, err := createSession(url, "pmm-server")
	require.NoError(t, err)

	err = sess.Database("test").Drop(context.TODO())
	require.NoError(t, err)

	ms := &testWriter{t: t}
	prof := New(url, logrus.WithField("component", "profiler-test"), ms, "test-id")
	err = prof.Start()
	defer prof.Stop()
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100)
		i := 0
		for i < int(aggregator.DefaultInterval*10) {
			fieldsCount := int(i/10) + 1
			doc := bson.M{}
			for j := 0; j < fieldsCount; j++ {
				doc[fmt.Sprintf("name_%05d", j)] = fmt.Sprintf("value_%05d", j)
			}
			<-ticker.C
			_, err = sess.Database("test").Collection("peoples").InsertOne(context.TODO(), doc)
			i++
		}
		wg.Done()
	}()

	//assert.NoError(t, err)

	<-time.After(aggregator.DefaultInterval)

	fmt.Printf("default interval: %v\n", aggregator.DefaultInterval)
	err = prof.Stop()
	require.NoError(t, err)
}

type testWriter struct {
	t *testing.T
}

func (tw *testWriter) Write(actual *report.Report) error {
	require.NotNil(tw.t, actual)
	assert.Equal(tw.t, 1, len(actual.Buckets))

	expected := &agentpb.MetricsBucket{
		Common: &agentpb.MetricsBucket_Common{
			Fingerprint: "INSERT peoples",
			Database:    "test",
			Schema:      "peoples",
			AgentId:     "test-id",
			AgentType:   inventorypb.AgentType_QAN_MONGODB_PROFILER_AGENT,
			NumQueries:  1,
		},
		Mongodb: &agentpb.MetricsBucket_MongoDB{
			MResponseLengthSum: 60,
			MResponseLengthMin: 60,
			MResponseLengthMax: 60,
		},
	}

	fmt.Println("====================================================================================================")
	pretty.Println(actual.Buckets)
	fmt.Println("====================================================================================================")
	assert.Equal(tw.t, expected, actual.Buckets[0])
	return nil
}
