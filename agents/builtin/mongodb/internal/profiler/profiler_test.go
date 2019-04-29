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

package profiler

import (
	"testing"
	"time"

	"github.com/percona/pmgo"
	"github.com/percona/pmm/api/qanpb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2/bson"

	"github.com/percona/pmm-agent/agents/builtin/mongodb/internal/profiler/aggregator"
	"github.com/percona/pmm-agent/agents/builtin/mongodb/internal/report"
	"github.com/percona/pmm-agent/agents/builtin/mongodb/internal/test/profiling"
)

func TestProfiler(t *testing.T) {
	defer func() {
		aggregator.DefaultInterval = time.Duration(time.Minute)
	}()
	aggregator.DefaultInterval = time.Duration(time.Second)
	url := "mongodb://pmm-agent:root-password@127.0.0.1:27017"
	p := profiling.New(url)

	dialInfo, err := pmgo.ParseURL(url)
	require.NoError(t, err)

	dialer := pmgo.NewDialer()

	sess, err := createSession(dialInfo, dialer)
	require.NoError(t, err)

	err = sess.DB("test").DropDatabase()
	require.NoError(t, err)

	err = p.Enable("test")
	require.NoError(t, err)

	ms := &testWriter{t: t}
	prof := New(dialInfo, dialer, logrus.WithField("component", "profiler-test"), ms, "test-id")
	err = prof.Start()
	require.NoError(t, err)

	err = sess.DB("test").C("peoples").Insert(bson.M{"name": "Anton"}, bson.M{"name": "Alexey"})
	require.NoError(t, err)

	<-time.After(aggregator.DefaultInterval)

	err = prof.Stop()
	require.NoError(t, err)

	err = p.Disable("test")
	require.NoError(t, err)
}

type testWriter struct {
	t *testing.T
}

func (tw *testWriter) Write(actual *report.Report) error {
	require.NotNil(tw.t, actual)
	assert.Equal(tw.t, 1, len(actual.Buckets))

	assert.Equal(tw.t, "INSERT peoples", actual.Buckets[0].Fingerprint)
	assert.Equal(tw.t, "test", actual.Buckets[0].DDatabase)
	assert.Equal(tw.t, "peoples", actual.Buckets[0].DSchema)
	assert.Equal(tw.t, "test-id", actual.Buckets[0].AgentId)
	assert.Equal(tw.t, qanpb.MetricsSource_MONGODB_PROFILER, actual.Buckets[0].MetricsSource)
	assert.Equal(tw.t, 1, actual.Buckets[0].NumQueries)
	assert.Equal(tw.t, 60, actual.Buckets[0].MResponseLengthSum)
	assert.Equal(tw.t, 60, actual.Buckets[0].MResponseLengthMin)
	assert.Equal(tw.t, 60, actual.Buckets[0].MResponseLengthMax)
	return nil
}
