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
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
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
	db, err := sql.Open("mysql", p.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var bytes []byte

	switch p.format {
	case ExplainFormatDefault:
		statement := fmt.Sprintf("EXPLAIN %s", p.query)
		out := explainRow{}
		r := db.QueryRowContext(ctx, statement)

		err = r.Scan(
			&out.Id,
			&out.SelectType,
			&out.Table,
			&out.Partitions,
			&out.Type,
			&out.PossibleKeys,
			&out.Key,
			&out.KeyLen,
			&out.Ref,
			&out.Rows,
			&out.Filtered,
			&out.Extra,
		)
		if err != nil {
			return nil, err
		}

		bytes, err = json.Marshal(out)
		if err != nil {
			return nil, err
		}
	case ExplainFormatJSON:
		statement := fmt.Sprintf("EXPLAIN FORMAT=JSON %s", p.query)
		out := explainRowJson{}

		r := db.QueryRowContext(ctx, statement)
		err = r.Scan(&out.Explain)
		if err != nil {
			return nil, err
		}

		bytes, err = json.Marshal(out)
		if err != nil {
			return nil, err
		}
	}

	return bytes, nil
}

type explainRow struct {
	Id           sql.NullInt64
	SelectType   sql.NullString
	Table        sql.NullString
	Partitions   sql.NullString
	Type         sql.NullString
	PossibleKeys sql.NullString
	Key          sql.NullString
	KeyLen       sql.NullString
	Ref          sql.NullString
	Rows         sql.NullInt64
	Filtered     sql.NullFloat64
	Extra        sql.NullString
}

type explainRowJson struct {
	Explain sql.NullString
}
