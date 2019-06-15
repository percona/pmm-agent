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

package perfschema

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/percona/pmm/api/inventorypb"
	"github.com/percona/pmm/api/qanpb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/mysql"

	"github.com/percona/pmm-agent/utils/tests"
)

func assertBucketsEqual(t *testing.T, expected, actual *qanpb.MetricsBucket) bool {
	t.Helper()
	return assert.Equal(t, proto.MarshalTextString(expected), proto.MarshalTextString(actual))
}

func TestPerfSchemaMakeBuckets(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		prev := map[string]*eventsStatementsSummaryByDigest{
			"Normal": {
				Digest:          pointer.ToString("Normal"),
				DigestText:      pointer.ToString("SELECT 'Normal'"),
				CountStar:       10,
				SumRowsAffected: 50,
			},
		}
		current := map[string]*eventsStatementsSummaryByDigest{
			"Normal": {
				Digest:          pointer.ToString("Normal"),
				DigestText:      pointer.ToString("SELECT 'Normal'"),
				CountStar:       15, // +5
				SumRowsAffected: 60, // +10
			},
		}
		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
		require.Len(t, actual, 1)
		expected := &qanpb.MetricsBucket{
			Queryid:          "Normal",
			Fingerprint:      "SELECT 'Normal'",
			AgentType:        inventorypb.AgentType_QAN_MYSQL_PERFSCHEMA_AGENT,
			NumQueries:       5,
			MRowsAffectedCnt: 5,
			MRowsAffectedSum: 10, // 60-50
		}
		assertBucketsEqual(t, expected, actual[0])
	})

	t.Run("New", func(t *testing.T) {
		prev := map[string]*eventsStatementsSummaryByDigest{}
		current := map[string]*eventsStatementsSummaryByDigest{
			"New": {
				Digest:          pointer.ToString("New"),
				DigestText:      pointer.ToString("SELECT 'New'"),
				CountStar:       10,
				SumRowsAffected: 50,
			},
		}
		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
		require.Len(t, actual, 1)
		expected := &qanpb.MetricsBucket{
			Queryid:          "New",
			Fingerprint:      "SELECT 'New'",
			AgentType:        inventorypb.AgentType_QAN_MYSQL_PERFSCHEMA_AGENT,
			NumQueries:       10,
			MRowsAffectedCnt: 10,
			MRowsAffectedSum: 50,
		}
		assertBucketsEqual(t, expected, actual[0])
	})

	t.Run("Same", func(t *testing.T) {
		prev := map[string]*eventsStatementsSummaryByDigest{
			"Same": {
				Digest:          pointer.ToString("Same"),
				DigestText:      pointer.ToString("SELECT 'Same'"),
				CountStar:       10,
				SumRowsAffected: 50,
			},
		}
		current := map[string]*eventsStatementsSummaryByDigest{
			"Same": {
				Digest:          pointer.ToString("Same"),
				DigestText:      pointer.ToString("SELECT 'Same'"),
				CountStar:       10,
				SumRowsAffected: 50,
			},
		}
		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
		require.Len(t, actual, 0)
	})

	t.Run("Truncate", func(t *testing.T) {
		prev := map[string]*eventsStatementsSummaryByDigest{
			"Truncate": {
				Digest:          pointer.ToString("Truncate"),
				DigestText:      pointer.ToString("SELECT 'Truncate'"),
				CountStar:       10,
				SumRowsAffected: 50,
			},
		}
		current := map[string]*eventsStatementsSummaryByDigest{}
		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
		require.Len(t, actual, 0)
	})

	t.Run("TruncateAndNew", func(t *testing.T) {
		prev := map[string]*eventsStatementsSummaryByDigest{
			"TruncateAndNew": {
				Digest:          pointer.ToString("TruncateAndNew"),
				DigestText:      pointer.ToString("SELECT 'TruncateAndNew'"),
				CountStar:       10,
				SumRowsAffected: 50,
			},
		}
		current := map[string]*eventsStatementsSummaryByDigest{
			"TruncateAndNew": {
				Digest:          pointer.ToString("TruncateAndNew"),
				DigestText:      pointer.ToString("SELECT 'TruncateAndNew'"),
				CountStar:       5,
				SumRowsAffected: 25,
			},
		}
		actual := makeBuckets(current, prev, logrus.WithField("test", t.Name()))
		require.Len(t, actual, 1)
		expected := &qanpb.MetricsBucket{
			Queryid:          "TruncateAndNew",
			Fingerprint:      "SELECT 'TruncateAndNew'",
			AgentType:        inventorypb.AgentType_QAN_MYSQL_PERFSCHEMA_AGENT,
			NumQueries:       5,
			MRowsAffectedCnt: 5,
			MRowsAffectedSum: 25,
		}
		assertBucketsEqual(t, expected, actual[0])
	})
}

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

func setup(t *testing.T, db *reform.DB) *PerfSchema {
	t.Helper()

	_, err := db.Exec("TRUNCATE performance_schema.events_statements_history")
	require.NoError(t, err)
	_, err = db.Exec("TRUNCATE performance_schema.events_statements_summary_by_digest")
	require.NoError(t, err)

	return newPerfSchema(db, "agent_id", logrus.WithField("test", t.Name()))
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

func TestPerfSchema(t *testing.T) {
	sqlDB := tests.OpenTestMySQL(t)
	defer sqlDB.Close() //nolint:errcheck
	db := reform.NewDB(sqlDB, mysql.Dialect, reform.NewPrintfLogger(t.Logf))

	_, err := db.Exec("UPDATE performance_schema.setup_consumers SET ENABLED='YES' WHERE NAME='events_statements_history'")
	require.NoError(t, err, "failed to enable events_statements_history consumer")

	structs, err := db.SelectAllFrom(setupConsumersView, "ORDER BY NAME")
	require.NoError(t, err)
	logTable(t, structs)
	structs, err = db.SelectAllFrom(setupInstrumentsView, "ORDER BY NAME")
	require.NoError(t, err)
	logTable(t, structs)

	mySQLVersion, mySQLVendor := tests.MySQLVersion(t, sqlDB)
	var digests map[string]string // digest_text/fingerprint to digest/query_id
	switch fmt.Sprintf("%s-%s", mySQLVersion, mySQLVendor) {
	case "5.6-oracle":
		digests = map[string]string{
			"SELECT `sleep` (?)":   "192ad18c482d389f36ebb0aa58311236",
			"SELECT * FROM `city`": "cf5d7abca54943b1aa9e126c85a7d020",
		}
	case "5.7-oracle":
		digests = map[string]string{
			"SELECT `sleep` (?)":   "52f680b0d3b57c2fa381f52038754db4",
			"SELECT * FROM `city`": "05292e6e5fb868ce2864918d5e934cb3",
		}

	case "5.6-percona":
		digests = map[string]string{
			"SELECT `sleep` (?)":   "d8dc769e3126abd5578679f520bad1a5",
			"SELECT * FROM `city`": "6d3c8e264bfdd0ce5d3c81d481148a9c",
		}
	case "5.7-percona":
		digests = map[string]string{
			"SELECT `sleep` (?)":   "049a1b20acee144f86b9a1e4aca398d6",
			"SELECT * FROM `city`": "9c799bdb2460f79b3423b77cd10403da",
		}

	case "8.0-oracle", "8.0-percona": // Percona switched to upstream's implementation
		digests = map[string]string{
			"SELECT `sleep` (?)":   "0b1b1c39d4ee2dda7df2a532d0a23406d86bd34e2cd7f22e3f7e9dedadff9b69",
			"SELECT * FROM `city`": "950bdc225cf73c9096ba499351ed4376f4526abad3d8ceabc168b6b28cfc9eab",
		}

	case "10.2-mariadb":
		digests = map[string]string{
			"SELECT `sleep` (?)":   "e58c348e4947db23b7f3ad30b7ed184a",
			"SELECT * FROM `city`": "e0f47172152e8750d070a854e607123f",
		}

	case "10.3-mariadb":
		digests = map[string]string{
			"SELECT `sleep` (?)":   "af50128de9089f71d749eda5ba3d02cd",
			"SELECT * FROM `city`": "2153d686f335a2ca39f3aca05bf9709a",
		}

	case "10.4-mariadb":
		digests = map[string]string{
			"SELECT `sleep` (?)":   "84a33aa2dff8b023bfd9c28247516e55",
			"SELECT * FROM `city`": "639b3ffc239a110c57ade746773952ab",
		}

	default:
		t.Log("Unhandled version, assuming dummy digests.")
		digests = map[string]string{
			"SELECT `sleep` (?)":   "TODO-sleep",
			"SELECT * FROM `city`": "TODO-star",
		}
	}

	t.Run("Sleep", func(t *testing.T) {
		m := setup(t, db)

		_, err := db.Exec("SELECT /* Sleep */ sleep(0.1)")
		require.NoError(t, err)

		require.NoError(t, m.refreshHistoryCache())

		buckets, err := m.getNewBuckets(time.Date(2019, 4, 1, 10, 59, 0, 0, time.UTC), 60)
		require.NoError(t, err)
		buckets = filter(buckets)
		require.Len(t, buckets, 1)

		actual := buckets[0]
		assert.InDelta(t, 0.1, actual.MQueryTimeSum, 0.09)
		expected := &qanpb.MetricsBucket{
			Fingerprint:         "SELECT `sleep` (?)",
			Schema:              "world",
			AgentId:             "agent_id",
			PeriodStartUnixSecs: 1554116340,
			PeriodLengthSecs:    60,
			AgentType:           inventorypb.AgentType_QAN_MYSQL_PERFSCHEMA_AGENT,
			Example:             "SELECT /* Sleep */ sleep(0.1)",
			ExampleFormat:       qanpb.ExampleFormat_EXAMPLE,
			ExampleType:         qanpb.ExampleType_RANDOM,
			NumQueries:          1,
			MQueryTimeCnt:       1,
			MQueryTimeSum:       actual.MQueryTimeSum,
			MRowsSentCnt:        1,
			MRowsSentSum:        1,
		}
		expected.Queryid = digests[expected.Fingerprint]
		assertBucketsEqual(t, expected, actual)
	})

	t.Run("AllCities", func(t *testing.T) {
		m := setup(t, db)

		_, err := db.Exec("SELECT /* AllCities */ * FROM city")
		require.NoError(t, err)

		require.NoError(t, m.refreshHistoryCache())

		buckets, err := m.getNewBuckets(time.Date(2019, 4, 1, 10, 59, 0, 0, time.UTC), 60)
		require.NoError(t, err)
		buckets = filter(buckets)
		require.Len(t, buckets, 1)

		actual := buckets[0]
		assert.InDelta(t, 0, actual.MQueryTimeSum, 0.09)
		assert.InDelta(t, 0, actual.MLockTimeSum, 0.09)
		expected := &qanpb.MetricsBucket{
			Fingerprint:         "SELECT * FROM `city`",
			Schema:              "world",
			AgentId:             "agent_id",
			PeriodStartUnixSecs: 1554116340,
			PeriodLengthSecs:    60,
			AgentType:           inventorypb.AgentType_QAN_MYSQL_PERFSCHEMA_AGENT,
			Example:             "SELECT /* AllCities */ * FROM city",
			ExampleFormat:       qanpb.ExampleFormat_EXAMPLE,
			ExampleType:         qanpb.ExampleType_RANDOM,
			NumQueries:          1,
			MQueryTimeCnt:       1,
			MQueryTimeSum:       actual.MQueryTimeSum,
			MLockTimeCnt:        1,
			MLockTimeSum:        actual.MLockTimeSum,
			MRowsSentCnt:        1,
			MRowsSentSum:        4079,
			MRowsExaminedCnt:    1,
			MRowsExaminedSum:    4079,
			MFullScanCnt:        1,
			MFullScanSum:        1,
			MNoIndexUsedCnt:     1,
			MNoIndexUsedSum:     1,
		}
		expected.Queryid = digests[expected.Fingerprint]
		assertBucketsEqual(t, expected, actual)
	})
}
