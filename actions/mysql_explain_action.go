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
)

type mysqlExplainOutputFormat string

var (
	ExplainFormatDefault mysqlExplainOutputFormat = "default"
	ExplainFormatJSON    mysqlExplainOutputFormat = "json"
)

type mysqlExplainAction struct {
	id     string
	dsn    string
	format mysqlExplainOutputFormat
	query  string
}

// NewMySQLExplainAction creates MySQL Explain Action.
// This is an Action that can run `EXPLAIN` command on MySQL service with given DSN.
func NewMySQLExplainAction(id, dsn, query string, format mysqlExplainOutputFormat) Action {
	return &mysqlExplainAction{
		id:     id,
		dsn:    dsn,
		format: format,
		query:  query,
	}
}

// ID returns unique Action id.
func (p *mysqlExplainAction) ID() string {
	return p.id
}

// Type returns Action name as as string.
func (p *mysqlExplainAction) Type() string {
	return "mysql-explain"
}

// Run starts an Action. This method is blocking.
func (p *mysqlExplainAction) Run(ctx context.Context) ([]byte, error) {
	return nil, nil
}
