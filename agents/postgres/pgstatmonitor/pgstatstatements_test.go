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

package pgstatmonitor

import (
	"fmt"
	"strings"
	"testing"

	"github.com/percona/pmm/api/agentpb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/postgresql"

	"github.com/percona/pmm-agent/utils/tests"
)

func setup(t *testing.T, db *reform.DB) *PGStatMonitorQAN {
	t.Helper()

	selectQuery := fmt.Sprintf("SELECT /* %s */ ", queryTag) //nolint:gosec

	_, err := db.Exec(selectQuery + "pg_stat_monitor_reset()")
	require.NoError(t, err)

	return newPgStatMonitorQAN(db.WithTag(queryTag), nil, "agent_id", logrus.WithField("test", t.Name()))
}

// filter removes buckets for queries that are not expected by tests.
func filter(mb []*agentpb.MetricsBucket) []*agentpb.MetricsBucket {
	res := make([]*agentpb.MetricsBucket, 0, len(mb))
	for _, b := range mb {
		switch {
		case strings.Contains(b.Common.Fingerprint, "/* pmm-agent:pgstatmonitor */"):
			continue
		default:
			res = append(res, b)
		}
	}
	return res
}

func TestPGStatMonitorQAN(t *testing.T) {
	sqlDB := tests.OpenTestPostgreSQL(t)
	defer sqlDB.Close() //nolint:errcheck
	db := reform.NewDB(sqlDB, postgresql.Dialect, reform.NewPrintfLogger(t.Logf))

	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_stat_monitor SCHEMA public")
	require.NoError(t, err)

}
