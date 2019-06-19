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

package pgstatsstatements

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/percona/pmm/api/inventorypb"
	"github.com/percona/pmm/api/qanpb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/postgresql"

	"github.com/percona/pmm-agent/utils/tests"
)

func assertBucketsEqual(t *testing.T, expected, actual *qanpb.MetricsBucket) bool {
	t.Helper()
	return assert.Equal(t, proto.MarshalTextString(expected), proto.MarshalTextString(actual))
}

//
//func TestPerfSchemaMakeBuckets(t *testing.T) {
//	t.Run("Normal", func(t *testing.T) {
//		prev := map[string]*pgStatStatements{
//			"1": {
//				Queryid: pointer.ToInt64(1),
//				Query:   pointer.ToString("SELECT 'Normal'"),
//				Calls:   pointer.ToInt64(10),
//			},
//		}
//		current := map[string]*pgStatStatements{
//			"1": {
//				Queryid: pointer.ToInt64(1),
//				Query:   pointer.ToString("SELECT 'Normal'"),
//				Calls:   pointer.ToInt64(10),
//			},
//		}
//		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
//		require.Len(t, actual, 1)
//		expected := &qanpb.MetricsBucket{
//			Queryid:          "Normal",
//			Fingerprint:      "SELECT 'Normal'",
//			AgentType:        inventorypb.AgentType_QAN_MYSQL_PERFSCHEMA_AGENT,
//			NumQueries:       5,
//			MRowsAffectedCnt: 5,
//			MRowsAffectedSum: 10, // 60-50
//		}
//		assertBucketsEqual(t, expected, actual[0])
//	})
//
//	t.Run("New", func(t *testing.T) {
//		prev := map[string]*pgStatStatements{}
//		current := map[string]*pgStatStatements{
//			"New": {
//				Digest:          pointer.ToString("New"),
//				DigestText:      pointer.ToString("SELECT 'New'"),
//				CountStar:       10,
//				SumRowsAffected: 50,
//			},
//		}
//		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
//		require.Len(t, actual, 1)
//		expected := &qanpb.MetricsBucket{
//			Queryid:          "New",
//			Fingerprint:      "SELECT 'New'",
//			AgentType:        inventorypb.AgentType_QAN_MYSQL_PERFSCHEMA_AGENT,
//			NumQueries:       10,
//			MRowsAffectedCnt: 10,
//			MRowsAffectedSum: 50,
//		}
//		assertBucketsEqual(t, expected, actual[0])
//	})
//
//	t.Run("Same", func(t *testing.T) {
//		prev := map[string]*pgStatStatements{
//			"Same": {
//				Digest:          pointer.ToString("Same"),
//				DigestText:      pointer.ToString("SELECT 'Same'"),
//				CountStar:       10,
//				SumRowsAffected: 50,
//			},
//		}
//		current := map[string]*pgStatStatements{
//			"Same": {
//				Digest:          pointer.ToString("Same"),
//				DigestText:      pointer.ToString("SELECT 'Same'"),
//				CountStar:       10,
//				SumRowsAffected: 50,
//			},
//		}
//		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
//		require.Len(t, actual, 0)
//	})
//
//	t.Run("Truncate", func(t *testing.T) {
//		prev := map[string]*pgStatStatements{
//			"Truncate": {
//				Digest:          pointer.ToString("Truncate"),
//				DigestText:      pointer.ToString("SELECT 'Truncate'"),
//				CountStar:       10,
//				SumRowsAffected: 50,
//			},
//		}
//		current := map[string]*pgStatStatements{}
//		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
//		require.Len(t, actual, 0)
//	})
//
//	t.Run("TruncateAndNew", func(t *testing.T) {
//		prev := map[string]*pgStatStatements{
//			"TruncateAndNew": {
//				Digest:          pointer.ToString("TruncateAndNew"),
//				DigestText:      pointer.ToString("SELECT 'TruncateAndNew'"),
//				CountStar:       10,
//				SumRowsAffected: 50,
//			},
//		}
//		current := map[string]*pgStatStatements{
//			"TruncateAndNew": {
//				Digest:          pointer.ToString("TruncateAndNew"),
//				DigestText:      pointer.ToString("SELECT 'TruncateAndNew'"),
//				CountStar:       5,
//				SumRowsAffected: 25,
//			},
//		}
//		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
//		require.Len(t, actual, 1)
//		expected := &qanpb.MetricsBucket{
//			Queryid:          "TruncateAndNew",
//			Fingerprint:      "SELECT 'TruncateAndNew'",
//			AgentType:        inventorypb.AgentType_QAN_MYSQL_PERFSCHEMA_AGENT,
//			NumQueries:       5,
//			MRowsAffectedCnt: 5,
//			MRowsAffectedSum: 25,
//		}
//		assertBucketsEqual(t, expected, actual[0])
//	})
//}

func logTable(t *testing.T, structs []reform.Struct) {
	t.Helper()

	if len(structs) == 0 {
		t.Log("logTable: empty")
		return
	}

	columns := structs[0].View().Columns()
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.Debug)
	_, err := fmt.Fprintln(w, strings.Join(columns, "\t"))
	require.NoError(t, err)
	for i, c := range columns {
		columns[i] = strings.Repeat("-", len(c))
	}
	_, err = fmt.Fprintln(w, strings.Join(columns, "\t"))
	require.NoError(t, err)

	for _, str := range structs {
		res := make([]string, len(str.Values()))
		for i, v := range str.Values() {
			res[i] = spew.Sprint(v)
		}
		fmt.Fprintf(w, "%s\n", strings.Join(res, "\t"))
	}

	require.NoError(t, w.Flush())
	t.Logf("%s:\n%s", structs[0].View().Name(), buf.Bytes())
}

func setup(t *testing.T, db *reform.DB) *PGStatStatementsQAN {
	t.Helper()

	_, err := db.Exec(`do $$
	begin
	perform pg_stat_statements_reset();
	end
	$$;`)
	require.NoError(t, err)

	return newPgStatsStatementsQAN(db, "agent_id", logrus.WithField("test", t.Name()))
}

// filter removes buckets for queries that are not expected by tests.
func filter(mb []*qanpb.MetricsBucket) []*qanpb.MetricsBucket {
	res := make([]*qanpb.MetricsBucket, 0, len(mb))
	for _, b := range mb {
		switch {
		// actions tests, MySQLVersion helper
		case strings.HasPrefix(b.Fingerprint, "SHOW "):
			continue
		case strings.HasPrefix(b.Fingerprint, "ANALYZE "):
			continue
		case strings.HasPrefix(b.Fingerprint, "EXPLAIN "):
			continue

		// slowlog tests
		case strings.HasPrefix(b.Fingerprint, "SELECT @@`slow_query_"):
			continue

		case strings.HasPrefix(b.Fingerprint, "SELECT @@`skip_networking`"):
			continue

		case strings.HasPrefix(b.Fingerprint, "TRUNCATE `performance_schema`"):
			continue
		case strings.HasPrefix(b.Fingerprint, "SELECT `performance_schema`"):
			continue

		default:
			res = append(res, b)
		}
	}
	return res
}

func TestPGStatStatementsQAN(t *testing.T) {
	sqlDB := tests.OpenTestPostgreSQL(t)
	defer sqlDB.Close() //nolint:errcheck
	db := reform.NewDB(sqlDB, postgresql.Dialect, reform.NewPrintfLogger(t.Logf))

	structs, err := db.SelectAllFrom(pgStatDatabaseView, "")
	require.NoError(t, err)
	logTable(t, structs)
	structs, err = db.SelectAllFrom(pgStatStatementsView, "")
	require.NoError(t, err)
	logTable(t, structs)

	engineVersionMajor, engineVersionMinor, engine := tests.PostgreSQLVersion(t, sqlDB)
	var digests map[string]string // digest_text/fingerprint to digest/query_id
	switch fmt.Sprintf("%s-%s", engineVersionMajor, engine) {
	case "11-PostgreSQL":
		digests = map[string]string{
			"SELECT * FROM city": "-6046499049124467328",
		}
	case "9-PostgreSQL":
		switch engineVersionMinor {
		case "4":
			digests = map[string]string{
				"SELECT * FROM city": "2500439221",
			}
		case "5", "6":
			digests = map[string]string{
				"SELECT * FROM city": "3778117319",
			}
		}
	case "10-PostgreSQL":
		digests = map[string]string{
			"SELECT * FROM city": "952213449",
		}

	default:
		t.Log("Unhandled version, assuming dummy digests.")
		digests = map[string]string{
			"SELECT * FROM city": "-6046499049124467328",
		}
	}

	t.Run("AllCities", func(t *testing.T) {
		m := setup(t, db)

		_, err := db.Exec("SELECT * FROM city")
		require.NoError(t, err)

		buckets, err := m.getNewBuckets(time.Date(2019, 4, 1, 10, 59, 0, 0, time.UTC), 60)
		require.NoError(t, err)
		buckets = filter(buckets)
		require.Len(t, buckets, 3)

		var actual *qanpb.MetricsBucket
		for _, v := range buckets {
			if v.Fingerprint == "SELECT * FROM city" {
				actual = v
			}
		}
		assert.InDelta(t, 0, actual.MQueryTimeSum, 0.09)
		//assert.InDelta(t, 0, actual.MLockTimeSum, 0.09)
		expected := &qanpb.MetricsBucket{
			Fingerprint:         "SELECT * FROM city",
			Schema:              "pmm-agent",
			AgentId:             "agent_id",
			PeriodStartUnixSecs: 1554116340,
			PeriodLengthSecs:    60,
			AgentType:           inventorypb.AgentType_QAN_POSTGRESQL_PGSTATEMENTS_AGENT,
			//Example:             "SELECT /* AllCities */ * FROM city",
			//ExampleFormat:       qanpb.ExampleFormat_EXAMPLE,
			//ExampleType:         qanpb.ExampleType_RANDOM,
			NumQueries:    1,
			MQueryTimeCnt: 1,
			MQueryTimeSum: actual.MQueryTimeSum,
			//MLockTimeCnt:        1,
			//MLockTimeSum:        actual.MLockTimeSum,
			MRowsSentCnt: 1,
			MRowsSentSum: 4079,
			//MRowsExaminedCnt:    1,
			//MRowsExaminedSum:    4079,
			//MFullScanCnt:        1,
			//MFullScanSum:        1,
			//MNoIndexUsedCnt:     1,
			//MNoIndexUsedSum:     1,
		}
		expected.Queryid = digests[expected.Fingerprint]
		assertBucketsEqual(t, expected, actual)
	})
}
