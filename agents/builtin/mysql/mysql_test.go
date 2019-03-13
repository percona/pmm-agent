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

package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/mysql"

	"github.com/percona/pmm-agent/utils/tests"
)

func TestGet(t *testing.T) {
	sqlDB := tests.OpenTestMySQL(t)
	db := reform.NewDB(sqlDB, mysql.Dialect, reform.NewPrintfLogger(t.Logf))
	m := New(nil, nil)

	_, err := db.Exec("TRUNCATE performance_schema.events_statements_summary_by_digest")
	require.NoError(t, err)

	_, err = db.Exec("SELECT 'TestGet'")
	require.NoError(t, err)

	actual, err := m.get(db.Querier)
	require.NoError(t, err)
	assert.Len(t, actual.MetricsBucket, 2)
}
