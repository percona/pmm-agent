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

func TestMySQLQueryShow(t *testing.T) {
	t.Parallel()

	dsn := tests.GetTestMySQLDSN(t)
	db := tests.OpenTestMySQL(t)
	defer db.Close() //nolint:errcheck

	t.Run("Default", func(t *testing.T) {
		params := &agentpb.StartActionRequest_MySQLQueryShowParams{
			Dsn:   dsn,
			Query: "VARIABLES",
		}
		a := NewMySQLQueryShowAction("", params)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		b, err := a.Run(ctx)
		require.NoError(t, err)
		assert.LessOrEqual(t, 16345, len(b))
		assert.LessOrEqual(t, len(b), 21942)

		data, err := agentpb.UnmarshalActionQueryResult(b)
		require.NoError(t, err)
		t.Log(spew.Sdump(data))
		assert.LessOrEqual(t, 456, len(data))
		assert.LessOrEqual(t, len(data), 589)

		var found int
		for _, m := range data {
			value := m["Value"]
			switch string(m["Variable_name"].([]byte)) {
			case "auto_generate_certs":
				assert.Equal(t, []byte("auto_generate_certs"), value)
				found++
			case "auto_increment_increment":
				assert.Equal(t, []byte("1"), value)
				found++
			}
		}
		assert.Equal(t, 1, found)
	})
}
