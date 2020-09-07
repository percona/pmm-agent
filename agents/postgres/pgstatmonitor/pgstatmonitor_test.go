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
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/api/inventorypb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/postgresql"

	"github.com/percona/pmm-agent/utils/tests"
)

func setup(t *testing.T, db *reform.DB, disableQueryExamples bool) *PGStatMonitorQAN {
	t.Helper()

	selectQuery := fmt.Sprintf("SELECT /* %s */ ", queryTag) //nolint:gosec
	_, err := db.Exec(selectQuery + "pg_stat_monitor_reset()")
	require.NoError(t, err)

	pgStatMonitorQAN, err := newPgStatMonitorQAN(db.WithTag(queryTag), nil, "agent_id", disableQueryExamples, logrus.WithField("test", t.Name()))
	require.NoError(t, err)

	return pgStatMonitorQAN
}

func supportedVersion(version string) bool {
	supported := float64(11)
	current, err := strconv.ParseFloat(version, 32)
	if err != nil {
		return false
	}

	return current >= supported
}

func extensionExists(db *reform.DB) bool {
	var name string
	err := db.QueryRow("SELECT name FROM pg_available_extensions WHERE name='pg_stat_monitor'").Scan(&name)
	if err != nil {
		return false
	}

	return true
}

// filter removes buckets for queries that are not expected by tests.
func filter(mb []*agentpb.MetricsBucket) []*agentpb.MetricsBucket {
	res := make([]*agentpb.MetricsBucket, 0, len(mb))
	for _, b := range mb {
		switch {
		case strings.Contains(b.Common.Fingerprint, "/* pmm-agent:pgstatmonitor */"):
			continue
		case strings.Contains(b.Common.Example, "/* pmm-agent:pgstatmonitor */"):
			continue
		case strings.Contains(b.Common.Fingerprint, "pg_stat_monitor_settings"):
			continue
		case strings.Contains(b.Common.Example, "pg_stat_monitor_settings"):
			continue
		case strings.Contains(b.Common.Fingerprint, "pg_stat_monitor_reset()"):
			continue
		case strings.Contains(b.Common.Example, "pg_stat_monitor_reset()"):
			continue
		default:
			res = append(res, b)
		}
	}
	return res
}

func TestPGStatMonitorSchema(t *testing.T) {
	sqlDB := tests.OpenTestPostgreSQL(t)
	defer sqlDB.Close() //nolint:errcheck
	db := reform.NewDB(sqlDB, postgresql.Dialect, reform.NewPrintfLogger(t.Logf))

	engineVersion := tests.PostgreSQLVersion(t, sqlDB)
	if !supportedVersion(engineVersion) || !extensionExists(db) {
		t.Skip()
	}

	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_stat_monitor SCHEMA public")
	assert.NoError(t, err)

	structs, err := db.SelectAllFrom(pgStatMonitorView, "")
	require.NoError(t, err)
	tests.LogTable(t, structs)

	const selectAllCities = "SELECT /* AllCities */ * FROM city"
	const selectAllCitiesLong = "SELECT /* AllCitiesTruncated */ * FROM city WHERE id IN " +
		"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, " +
		"$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, " +
		"$41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60, " +
		"$61, $62, $63, $64, $65, $66, $67, $68, $69, $70, $71, $72, $73, $74, $75, $76, $77, $78, $79, $80, " +
		"$81, $82, $83, $84, $85, $86, $87, $88, $89, $90, $91, $92, $93, $94, $95, $96, $97, $98, $99, $100, " +
		"$101, $102, $103, $104, $105, $106, $107, $108, $109, $110, $111, $112, $113, $114, $115, $116, $117, $118, $119, $120, " +
		"$121, $122, $123, $124, $125, $126, $127, $128, $129, $130, $131, $132, $133, $134, $135, $136, $137, $138, $139, $140, " +
		"$141, $142, $143, $144, $145, $146, $147, $148, $149, $150, $151, $152, $153, $154, $155, $156, $157, $158, $159, $160, " +
		"$161, $162, $163, $164, $165, $166, $167, $168, $169, $170, $171, $172, $173, $174, $175, $176, $177, $178, $179, $"

	var digests map[string]string
	switch engineVersion {
	case "11":
		digests = map[string]string{
			selectAllCities:     "C7B4B1F338ABF1FF",
			selectAllCitiesLong: "1EFF7C2B26084540",
		}
	case "12":
		digests = map[string]string{
			selectAllCities:     "4E18B291CCDEC5E3",
			selectAllCitiesLong: "E9B974F0FBC9ED4A",
		}

	default:
		t.Log("Unhandled version, assuming dummy digests.")
		digests = map[string]string{
			selectAllCities:     "TODO-selectAllCities",
			selectAllCitiesLong: "TODO-selectAllCitiesLong",
		}
	}

	t.Run("AllCities", func(t *testing.T) {
		m := setup(t, db, true)

		_, err := db.Exec(selectAllCities)
		require.NoError(t, err)

		buckets, err := m.getNewBuckets(context.Background(), time.Date(2019, 4, 1, 10, 59, 0, 0, time.UTC), 60)
		require.NoError(t, err)
		buckets = filter(buckets)
		t.Logf("Actual:\n%s", tests.FormatBuckets(buckets))
		require.Len(t, buckets, 1)

		actual := buckets[0]
		assert.InDelta(t, 0, actual.Common.MQueryTimeSum, 0.09)
		assert.Equal(t, float32(32), actual.Postgresql.MSharedBlksHitSum+actual.Postgresql.MSharedBlksReadSum)
		assert.InDelta(t, 1.5, actual.Postgresql.MSharedBlksHitCnt+actual.Postgresql.MSharedBlksReadCnt, 0.5)
		expected := &agentpb.MetricsBucket{
			Common: &agentpb.MetricsBucket_Common{
				Fingerprint:         selectAllCities,
				Database:            "pmm-agent",
				Tables:              []string{"public.city"},
				Username:            "pmm-agent",
				AgentId:             "agent_id",
				PeriodStartUnixSecs: 1554116340,
				PeriodLengthSecs:    60,
				AgentType:           inventorypb.AgentType_QAN_POSTGRESQL_PGSTATMONITOR_AGENT,
				NumQueries:          1,
				MQueryTimeCnt:       1,
				MQueryTimeSum:       actual.Common.MQueryTimeSum,
			},
			Postgresql: &agentpb.MetricsBucket_PostgreSQL{
				MBlkReadTimeCnt:    actual.Postgresql.MBlkReadTimeCnt,
				MBlkReadTimeSum:    actual.Postgresql.MBlkReadTimeSum,
				MSharedBlksReadCnt: actual.Postgresql.MSharedBlksReadCnt,
				MSharedBlksReadSum: actual.Postgresql.MSharedBlksReadSum,
				MSharedBlksHitCnt:  actual.Postgresql.MSharedBlksHitCnt,
				MSharedBlksHitSum:  actual.Postgresql.MSharedBlksHitSum,
				MRowsCnt:           1,
				MRowsSum:           4079,
				MCpuUserTimeCnt:    actual.Postgresql.MCpuUserTimeCnt,
				MCpuUserTimeSum:    actual.Postgresql.MCpuUserTimeSum,
				MCpuSysTimeCnt:     actual.Postgresql.MCpuSysTimeCnt,
				MCpuSysTimeSum:     actual.Postgresql.MCpuSysTimeSum,
			},
		}
		expected.Common.Queryid = digests[expected.Common.Fingerprint]
		tests.AssertBucketsEqual(t, expected, actual)
		assert.LessOrEqual(t, actual.Postgresql.MBlkReadTimeSum, actual.Common.MQueryTimeSum)

		_, err = db.Exec(selectAllCities)
		require.NoError(t, err)

		buckets, err = m.getNewBuckets(context.Background(), time.Date(2019, 4, 1, 10, 59, 0, 0, time.UTC), 60)
		require.NoError(t, err)
		buckets = filter(buckets)
		t.Logf("Actual:\n%s", tests.FormatBuckets(buckets))
		require.Len(t, buckets, 1)

		actual = buckets[0]
		assert.InDelta(t, 0, actual.Common.MQueryTimeSum, 0.09)
		expected = &agentpb.MetricsBucket{
			Common: &agentpb.MetricsBucket_Common{
				Fingerprint:         selectAllCities,
				Database:            "pmm-agent",
				Tables:              []string{"public.city"},
				Username:            "pmm-agent",
				AgentId:             "agent_id",
				PeriodStartUnixSecs: 1554116340,
				PeriodLengthSecs:    60,
				AgentType:           inventorypb.AgentType_QAN_POSTGRESQL_PGSTATMONITOR_AGENT,
				NumQueries:          1,
				MQueryTimeCnt:       1,
				MQueryTimeSum:       actual.Common.MQueryTimeSum,
			},
			Postgresql: &agentpb.MetricsBucket_PostgreSQL{
				MSharedBlksHitCnt: 1,
				MSharedBlksHitSum: 32,
				MRowsCnt:          1,
				MRowsSum:          4079,
				MBlkReadTimeCnt:   actual.Postgresql.MBlkReadTimeCnt,
				MBlkReadTimeSum:   actual.Postgresql.MBlkReadTimeSum,
				MCpuUserTimeCnt:   actual.Postgresql.MCpuUserTimeCnt,
				MCpuUserTimeSum:   actual.Postgresql.MCpuUserTimeSum,
				MCpuSysTimeCnt:    actual.Postgresql.MCpuSysTimeCnt,
				MCpuSysTimeSum:    actual.Postgresql.MCpuSysTimeSum,
			},
		}
		expected.Common.Queryid = digests[expected.Common.Fingerprint]
		tests.AssertBucketsEqual(t, expected, actual)
		assert.LessOrEqual(t, actual.Postgresql.MBlkReadTimeSum, actual.Common.MQueryTimeSum)
	})

	t.Run("AllCitiesTruncated", func(t *testing.T) {
		m := setup(t, db, false)

		const n = 500
		placeholders := db.Placeholders(1, n)
		args := make([]interface{}, n)
		for i := 0; i < n; i++ {
			args[i] = i
		}
		q := fmt.Sprintf("SELECT /* AllCitiesTruncated */ * FROM city WHERE id IN (%s)", strings.Join(placeholders, ", ")) //nolint:gosec
		_, err := db.Exec(q, args...)
		require.NoError(t, err)

		buckets, err := m.getNewBuckets(context.Background(), time.Date(2019, 4, 1, 10, 59, 0, 0, time.UTC), 60)
		require.NoError(t, err)
		buckets = filter(buckets)
		t.Logf("Actual:\n%s", tests.FormatBuckets(buckets))
		require.Len(t, buckets, 1)

		actual := buckets[0]
		assert.InDelta(t, 0, actual.Common.MQueryTimeSum, 0.09)
		assert.InDelta(t, 1010, actual.Postgresql.MSharedBlksHitSum+actual.Postgresql.MSharedBlksReadSum, 3)
		assert.InDelta(t, 1.5, actual.Postgresql.MSharedBlksHitCnt+actual.Postgresql.MSharedBlksReadCnt, 0.5)
		expected := &agentpb.MetricsBucket{
			Common: &agentpb.MetricsBucket_Common{
				Fingerprint:         selectAllCitiesLong,
				Database:            "pmm-agent",
				Tables:              []string{"public.city"},
				Username:            "pmm-agent",
				AgentId:             "agent_id",
				PeriodStartUnixSecs: 1554116340,
				PeriodLengthSecs:    60,
				AgentType:           inventorypb.AgentType_QAN_POSTGRESQL_PGSTATMONITOR_AGENT,
				NumQueries:          1,
				MQueryTimeCnt:       1,
				MQueryTimeSum:       actual.Common.MQueryTimeSum,
			},
			Postgresql: &agentpb.MetricsBucket_PostgreSQL{
				MBlkReadTimeCnt:    actual.Postgresql.MBlkReadTimeCnt,
				MBlkReadTimeSum:    actual.Postgresql.MBlkReadTimeSum,
				MSharedBlksReadCnt: actual.Postgresql.MSharedBlksReadCnt,
				MSharedBlksReadSum: actual.Postgresql.MSharedBlksReadSum,
				MSharedBlksHitCnt:  actual.Postgresql.MSharedBlksHitCnt,
				MSharedBlksHitSum:  actual.Postgresql.MSharedBlksHitSum,
				MRowsCnt:           1,
				MRowsSum:           499,
				MCpuUserTimeCnt:    actual.Postgresql.MCpuUserTimeCnt,
				MCpuUserTimeSum:    actual.Postgresql.MCpuUserTimeSum,
				MCpuSysTimeCnt:     actual.Postgresql.MCpuSysTimeCnt,
				MCpuSysTimeSum:     actual.Postgresql.MCpuSysTimeSum,
			},
		}
		expected.Common.Queryid = digests[expected.Common.Fingerprint]
		tests.AssertBucketsEqual(t, expected, actual)
		assert.LessOrEqual(t, actual.Postgresql.MBlkReadTimeSum, actual.Common.MQueryTimeSum)

		_, err = db.Exec(q, args...)
		require.NoError(t, err)

		buckets, err = m.getNewBuckets(context.Background(), time.Date(2019, 4, 1, 10, 59, 0, 0, time.UTC), 60)
		require.NoError(t, err)
		buckets = filter(buckets)
		t.Logf("Actual:\n%s", tests.FormatBuckets(buckets))
		require.Len(t, buckets, 1)

		actual = buckets[0]
		assert.InDelta(t, 0, actual.Common.MQueryTimeSum, 0.09)
		assert.InDelta(t, 0, actual.Postgresql.MBlkReadTimeCnt, 1)
		assert.InDelta(t, 1007, actual.Postgresql.MSharedBlksHitSum, 2)
		expected = &agentpb.MetricsBucket{
			Common: &agentpb.MetricsBucket_Common{
				Fingerprint:         selectAllCitiesLong,
				Database:            "pmm-agent",
				Tables:              []string{"public.city"},
				Username:            "pmm-agent",
				AgentId:             "agent_id",
				PeriodStartUnixSecs: 1554116340,
				PeriodLengthSecs:    60,
				AgentType:           inventorypb.AgentType_QAN_POSTGRESQL_PGSTATMONITOR_AGENT,
				NumQueries:          1,
				MQueryTimeCnt:       1,
				MQueryTimeSum:       actual.Common.MQueryTimeSum,
			},
			Postgresql: &agentpb.MetricsBucket_PostgreSQL{
				MBlkReadTimeCnt:   actual.Postgresql.MBlkReadTimeCnt,
				MBlkReadTimeSum:   actual.Postgresql.MBlkReadTimeSum,
				MSharedBlksHitCnt: 1,
				MSharedBlksHitSum: actual.Postgresql.MSharedBlksHitSum,
				MRowsCnt:          1,
				MRowsSum:          499,
				MCpuUserTimeCnt:   actual.Postgresql.MCpuUserTimeCnt,
				MCpuUserTimeSum:   actual.Postgresql.MCpuUserTimeSum,
				MCpuSysTimeCnt:    actual.Postgresql.MCpuSysTimeCnt,
				MCpuSysTimeSum:    actual.Postgresql.MCpuSysTimeSum,
			},
		}
		expected.Common.Queryid = digests[expected.Common.Fingerprint]
		tests.AssertBucketsEqual(t, expected, actual)
		assert.LessOrEqual(t, actual.Postgresql.MBlkReadTimeSum, actual.Common.MQueryTimeSum)
	})

	t.Run("CheckMBlkReadTime", func(t *testing.T) {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		tableName := fmt.Sprintf("customer%d", r.Int())
		_, err := db.Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			customer_id integer NOT NULL,
			first_name character varying(45) NOT NULL,
			last_name character varying(45) NOT NULL,
			active boolean
		)`, tableName))
		require.NoError(t, err)
		defer func() {
			_, err := db.Exec(fmt.Sprintf(`DROP TABLE %s`, tableName))
			require.NoError(t, err)
		}()
		m := setup(t, db, true)

		var waitGroup sync.WaitGroup
		n := 1000
		for i := 0; i < n; i++ {
			id := i
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				_, err := db.Exec(fmt.Sprintf(`INSERT /* CheckMBlkReadTime */ INTO %s (customer_id, first_name, last_name, active) VALUES (%d, 'John', 'Dow', TRUE)`, tableName, id))
				require.NoError(t, err)
			}()
		}
		waitGroup.Wait()

		buckets, err := m.getNewBuckets(context.Background(), time.Date(2020, 5, 25, 10, 59, 0, 0, time.UTC), 60)
		require.NoError(t, err)
		buckets = filter(buckets)
		t.Logf("Actual:\n%s", tests.FormatBuckets(buckets))
		require.Len(t, buckets, 1)

		actual := buckets[0]
		assert.NotZero(t, actual.Postgresql.MBlkReadTimeSum)
		var expected = &agentpb.MetricsBucket{
			Common: &agentpb.MetricsBucket_Common{
				Queryid:             actual.Common.Queryid,
				Fingerprint:         fmt.Sprintf("INSERT /* CheckMBlkReadTime */ INTO %s (customer_id, first_name, last_name, active) VALUES ($1, $2, $3, $4)", tableName),
				Database:            "pmm-agent",
				Username:            "pmm-agent",
				AgentId:             "agent_id",
				PeriodStartUnixSecs: 1590404340,
				PeriodLengthSecs:    60,
				AgentType:           inventorypb.AgentType_QAN_POSTGRESQL_PGSTATMONITOR_AGENT,
				NumQueries:          float32(n),
				MQueryTimeCnt:       float32(n),
				MQueryTimeSum:       actual.Common.MQueryTimeSum,
			},
			Postgresql: &agentpb.MetricsBucket_PostgreSQL{
				MBlkReadTimeCnt:       float32(n),
				MBlkReadTimeSum:       actual.Postgresql.MBlkReadTimeSum,
				MSharedBlksReadCnt:    actual.Postgresql.MSharedBlksReadCnt,
				MSharedBlksReadSum:    actual.Postgresql.MSharedBlksReadSum,
				MSharedBlksWrittenCnt: actual.Postgresql.MSharedBlksWrittenCnt,
				MSharedBlksWrittenSum: actual.Postgresql.MSharedBlksWrittenSum,
				MSharedBlksDirtiedCnt: actual.Postgresql.MSharedBlksDirtiedCnt,
				MSharedBlksDirtiedSum: actual.Postgresql.MSharedBlksDirtiedSum,
				MSharedBlksHitCnt:     actual.Postgresql.MSharedBlksHitCnt,
				MSharedBlksHitSum:     actual.Postgresql.MSharedBlksHitSum,
				MRowsCnt:              float32(n),
				MRowsSum:              float32(n),
				MCpuUserTimeCnt:       actual.Postgresql.MCpuUserTimeCnt,
				MCpuUserTimeSum:       actual.Postgresql.MCpuUserTimeSum,
				MCpuSysTimeCnt:        actual.Postgresql.MCpuSysTimeCnt,
				MCpuSysTimeSum:        actual.Postgresql.MCpuSysTimeSum,
			},
		}
		tests.AssertBucketsEqual(t, expected, actual)
		assert.LessOrEqual(t, actual.Postgresql.MBlkReadTimeSum, actual.Common.MQueryTimeSum)
	})
}
