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

package actions

import (
	"context"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/percona/pmm/api/agentpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/pmm-agent/utils/tests"
)

func TestPostgreSQLQueryShow(t *testing.T) {
	t.Parallel()

	dsn := tests.GetTestPostgreSQLDSN(t)
	db := tests.OpenTestPostgreSQL(t)
	defer db.Close() //nolint:errcheck

	t.Run("Default", func(t *testing.T) {
		params := &agentpb.StartActionRequest_PostgreSQLQueryShowParams{
			Dsn: dsn,
		}
		a := NewPostgreSQLQueryShowAction("", params)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		b, err := a.Run(ctx)
		require.NoError(t, err)
		assert.LessOrEqual(t, 25235, len(b))
		assert.LessOrEqual(t, len(b), 26895)

		data, err := agentpb.UnmarshalActionQueryResult(b)
		require.NoError(t, err)
		t.Log(spew.Sdump(data))
		assert.LessOrEqual(t, 274, len(data))
		assert.LessOrEqual(t, len(data), 294)

		expected := map[string]interface{}{
			"name":        []byte("allow_system_table_mods"),
			"setting":     []byte("off"),
			"description": []byte("Allows modifications of the structure of system tables."),
		}
		assert.Equal(t, expected, data[0])

		expected = map[string]interface{}{
			"name":        []byte("autovacuum_freeze_max_age"),
			"setting":     []byte("200000000"),
			"description": []byte("Age at which to autovacuum a table to prevent transaction ID wraparound."),
		}
		assert.Equal(t, expected, data[10])
	})
}
