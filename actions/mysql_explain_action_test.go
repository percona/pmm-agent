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
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMysqlExplainActionRun(t *testing.T) {
	id := "/action_id/6a479303-5081-46d0-baa0-87d6248c987b"
	dsn := "pmm-agent:pmm-agent-password@tcp(127.0.0.1:3306)/information_schema"
	q := "SELECT * FROM information_schema.GLOBAL_STATUS"

	exp := NewMySQLExplainAction(id, dsn, q, ExplainFormatDefault)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := exp.Run(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, out)
	t.Log(string(out))
}

func TestMysqlExplainActionRunJson(t *testing.T) {
	id := "/action_id/6a479303-5081-46d0-baa0-87d6248c987b"
	dsn := "pmm-agent:pmm-agent-password@tcp(127.0.0.1:3306)/information_schema"
	q := "SELECT * FROM information_schema.GLOBAL_STATUS"

	exp := NewMySQLExplainAction(id, dsn, q, ExplainFormatJSON)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := exp.Run(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, out)
	t.Log(string(out))
}
