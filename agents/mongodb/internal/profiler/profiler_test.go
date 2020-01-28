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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/api/inventorypb"

	"github.com/percona/pmm-agent/agents/mongodb/internal/profiler/aggregator"
	"github.com/percona/pmm-agent/agents/mongodb/internal/report"
)

type MongoVersion struct {
	VersionString string `bson:"version"`
	PSMDBVersion  string `bson:"psmdbVersion"`
	Version       []int  `bson:"versionArray"`
}

func GetMongoVersion(ctx context.Context, client *mongo.Client) (string, error) {
	ver := new(MongoVersion)
	err := client.Database("admin").RunCommand(ctx, bson.D{{"buildInfo", 1}}).Decode(ver)
	if err != nil {
		return "", nil
	}

	version := fmt.Sprintf("%d.%d", ver.Version[0], ver.Version[1])
	return version, err
}

func TestProfiler(t *testing.T) {
	defaultInterval := aggregator.DefaultInterval
	aggregator.DefaultInterval = time.Duration(time.Second)
	defer func() { aggregator.DefaultInterval = defaultInterval }()

	url := "mongodb://root:root-password@127.0.0.1:27017"

	sess, err := createSession(url, "pmm-agent")
	require.NoError(t, err)

	cleanUpDBs(t, sess) // Just in case there are old dbs with matching names

	i := 0
	// It's done to create Databases.
	for i < 30 {
		doc := bson.M{"id": i}
		sess.Database(fmt.Sprintf("test_%02d", i)).Collection("people").InsertOne(context.TODO(), doc)
		i++
	}
	<-time.After(aggregator.DefaultInterval * 2) // give it some time before starting profiler

	ms := &testWriter{
		t:       t,
		reports: make([]*report.Report, 0),
	}
	prof := New(url, logrus.WithField("component", "profiler-test"), ms, "test-id")
	err = prof.Start()
	defer prof.Stop()
	require.NoError(t, err)
	<-time.After(aggregator.DefaultInterval) // give it some time to start profiler

	ticker := time.NewTicker(time.Millisecond * 1)
	i = 0
	for i < 300 {
		<-ticker.C
		fieldsCount := int(i/10) + 1
		doc := bson.M{}
		for j := 0; j < fieldsCount; j++ {
			doc[fmt.Sprintf("name_%05d", j)] = fmt.Sprintf("value_%05d", j) // to generate different fingerprints
		}
		_, err = sess.Database(fmt.Sprintf("test_%02d", int(i/10))).Collection("people").InsertOne(context.TODO(), doc)
		i++
	}
	<-time.After(aggregator.DefaultInterval * 3) // give it some time to catch all metrics

	err = prof.Stop()
	require.NoError(t, err)

	defer cleanUpDBs(t, sess)

	require.GreaterOrEqual(t, len(ms.reports), 1)

	var buckets []*agentpb.MetricsBucket
	for _, r := range ms.reports {
		fmt.Printf("%v\n", len(r.Buckets))
		buckets = append(buckets, r.Buckets...)
	}

	//buckets := ms.reports[0].Buckets
	sort.Slice(buckets, func(i, j int) bool { return buckets[i].Common.Database < buckets[j].Common.Database })

	version, err := GetMongoVersion(context.TODO(), sess)
	require.NoError(t, err)

	var responseLength float32
	switch version {
	case "3.4":
		responseLength = 44
	case "3.6":
		responseLength = 29
	default:
		fmt.Printf("%v", version)
		responseLength = 45
	}

	assert.Equal(t, 30, len(buckets)) // 300 sample docs / 10 = different database names
	for i := 0; i < len(buckets); i++ {
		assert.Equal(t, fmt.Sprintf("test_%02d", i), buckets[i].Common.Database)
		assert.Equal(t, "INSERT people", buckets[i].Common.Fingerprint)
		assert.Equal(t, []string{"people"}, buckets[i].Common.Tables)
		assert.Equal(t, "test-id", buckets[i].Common.AgentId)
		assert.Equal(t, inventorypb.AgentType(9), buckets[i].Common.AgentType)
		wantMongoDB := &agentpb.MetricsBucket_MongoDB{
			MDocsReturnedCnt:   10,
			MResponseLengthCnt: 10,
			MResponseLengthSum: responseLength * 10,
			MResponseLengthMin: responseLength,
			MResponseLengthMax: responseLength,
			MResponseLengthP99: responseLength,
			MDocsScannedCnt:    10,
		}
		assert.Equal(t, wantMongoDB, buckets[i].Mongodb)
	}
}

func cleanUpDBs(t *testing.T, sess *mongo.Client) {
	dbs, err := sess.ListDatabaseNames(context.TODO(), bson.M{})
	for _, dbname := range dbs {
		if strings.HasPrefix("test_", dbname) {
			err = sess.Database(dbname).Drop(context.TODO())
			require.NoError(t, err)
		}
	}
}

type testWriter struct {
	t       *testing.T
	reports []*report.Report
}

func (tw *testWriter) Write(actual *report.Report) error {
	require.NotNil(tw.t, actual)
	tw.reports = append(tw.reports, actual)
	return nil
}
