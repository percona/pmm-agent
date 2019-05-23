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

package actions

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLExplain(t *testing.T) {
	const query = "SELECT * FROM `city`"

	t.Run("Default", func(t *testing.T) {
		t.Parallel()

		dsn := "root:root-password@tcp(127.0.0.1:3306)/world"
		a := NewMySQLExplainAction("", dsn, query, ExplainFormatDefault)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		b, err := a.Run(ctx)
		require.NoError(t, err)
		actual := strings.TrimSpace(string(b))
		expected := strings.TrimSpace(`
id |select_type |table |type |possible_keys |key  |key_len |ref  |rows |Extra
1  |SIMPLE      |city  |ALL  |NULL          |NULL |NULL    |NULL |2    |NULL
	`)
		assert.Equal(t, expected, actual)
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()

		dsn := "root:root-password@tcp(127.0.0.1:3306)/world"
		a := NewMySQLExplainAction("", dsn, query, ExplainFormatJSON)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		b, err := a.Run(ctx)
		require.NoError(t, err)
		actual := strings.TrimSpace(string(b))
		expected := strings.TrimSpace(`
{
  "query_block": {
    "select_id": 1,
    "table": {
      "table_name": "city",
      "access_type": "ALL",
      "rows": 2,
      "filtered": 100
    }
  }
}
	`)
		assert.Equal(t, expected, actual)
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		dsn := "pmm-agent:pmm-agent-wrong-password@tcp(127.0.0.1:3306)/world"
		a := NewMySQLExplainAction("", dsn, query, ExplainFormatDefault)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err := a.Run(ctx)
		require.Error(t, err)
		assert.Regexp(t, `Error 1045: Access denied for user 'pmm-agent'@'.+' \(using password: YES\)`, err.Error())
	})
}
