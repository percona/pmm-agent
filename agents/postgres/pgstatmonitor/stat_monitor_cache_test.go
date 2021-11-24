package pgstatmonitor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/postgresql"

	"github.com/percona/pmm-agent/utils/tests"
)

func TestPGStatMonitorStructs(t *testing.T) {
	sqlDB := tests.OpenTestPostgreSQL(t)
	defer sqlDB.Close() //nolint:errcheck
	db := reform.NewDB(sqlDB, postgresql.Dialect, reform.NewPrintfLogger(t.Logf))

	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_stat_monitor SCHEMA public")
	assert.NoError(t, err)

	defer func() {
		_, err = db.Exec("DROP EXTENSION pg_stat_monitor")
		assert.NoError(t, err)
	}()

	engineVersion := tests.PostgreSQLVersion(t, sqlDB)
	if !supportedVersion(engineVersion) || !extensionExists(db) {
		t.Skip()
	}

	m := setup(t, db, false)
	current, cache, err := m.monitorCache.getStatMonitorExtended(context.TODO(), db.Querier, m.pgsmNormalizedQuery)

	require.NoError(t, err)
	require.NotNil(t, current)
	require.NotNil(t, cache)
}
