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

	"github.com/percona/pmm/api/agentpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/pmm-agent/utils/tests"
)

func TestPostgreSQLQuerySelect(t *testing.T) {
	t.Parallel()

	dsn := tests.GetTestPostgreSQLDSN(t)
	db := tests.OpenTestPostgreSQL(t)
	defer db.Close() //nolint:errcheck

	t.Run("Default", func(t *testing.T) {
		params := &agentpb.StartActionRequest_PostgreSQLQuerySelectParams{
			Dsn:   dsn,
			Query: "* FROM pg_extension",
		}
		a := NewPostgreSQLQuerySelectAction("", params)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		b, err := a.Run(ctx)
		require.NoError(t, err)
		assert.InDelta(t, 130, len(b), 1)

		data, err := agentpb.UnmarshalActionQueryResult(b)
		require.NoError(t, err)
		assert.InDelta(t, 1, len(data), 0)
		expected := map[string]interface{}{
			"extname":        []byte("plpgsql"), // name type
			"extowner":       []byte("10"),      // oid type
			"extnamespace":   []byte("11"),      // oid type
			"extrelocatable": false,
			"extversion":     "1.0", // text type
			"extconfig":      nil,
			"extcondition":   nil,
		}
		assert.Equal(t, expected, data[0])
	})

	t.Run("Binary", func(t *testing.T) {
		params := &agentpb.StartActionRequest_PostgreSQLQuerySelectParams{
			Dsn:   dsn,
			Query: `'\x0001feff'::bytea AS bytes`,
		}
		a := NewPostgreSQLQuerySelectAction("", params)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		b, err := a.Run(ctx)
		require.NoError(t, err)
		assert.InDelta(t, 17, len(b), 1)

		data, err := agentpb.UnmarshalActionQueryResult(b)
		require.NoError(t, err)
		assert.InDelta(t, 1, len(data), 0)
		expected := map[string]interface{}{
			"bytes": []byte{0x00, 0x01, 0xfe, 0xff},
		}
		assert.Equal(t, expected, data[0])
	})

	t.Run("LittleBobbyTables", func(t *testing.T) {
		params := &agentpb.StartActionRequest_PostgreSQLQuerySelectParams{
			Dsn:   dsn,
			Query: "* FROM city; DROP TABLE city CASCADE; --",
		}
		a := NewPostgreSQLQuerySelectAction("", params)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		b, err := a.Run(ctx)
		assert.EqualError(t, err, "query contains ';'")
		assert.Nil(t, b)

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM city").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 4079, count)
	})
}