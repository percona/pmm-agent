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
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"text/tabwriter"

	_ "github.com/go-sql-driver/mysql" // register SQL driver
	"github.com/pkg/errors"
)

// MysqlExplainOutputFormat explain output format.
type MysqlExplainOutputFormat int32

var (
	// ExplainFormatDefault default (table) explain format.
	ExplainFormatDefault MysqlExplainOutputFormat = 1
	// ExplainFormatJSON json explain format.
	ExplainFormatJSON MysqlExplainOutputFormat = 2
)

type nullString string

func (ns *nullString) String() string {
	if ns == nil {
		return "NULL"
	}
	return string(*ns)
}

type mysqlExplainAction struct {
	id     string
	dsn    string
	query  string
	format MysqlExplainOutputFormat
}

// NewMySQLExplainAction creates MySQL Explain Action.
// This is an Action that can run `EXPLAIN` command on MySQL service with given DSN.
func NewMySQLExplainAction(id, dsn, query string, format MysqlExplainOutputFormat) Action {
	return &mysqlExplainAction{
		id:     id,
		dsn:    dsn,
		query:  query,
		format: format,
	}
}

// ID returns an Action ID.
func (e *mysqlExplainAction) ID() string {
	return e.id
}

// Type returns an Action type.
func (e *mysqlExplainAction) Type() string {
	return "mysql-explain"
}

// Run runs an Action and returns output and error.
func (e *mysqlExplainAction) Run(ctx context.Context) ([]byte, error) {
	// TODO use ctx for connection

	db, err := sql.Open("mysql", e.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close() //nolint:errcheck

	switch e.format {
	case ExplainFormatDefault:
		return e.explain(ctx, db)
	case ExplainFormatJSON:
		return e.explainJSON(ctx, db)
	default:
		return nil, errors.New("unsupported output format")
	}
}

func (e *mysqlExplainAction) sealed() {}

func (e *mysqlExplainAction) explain(ctx context.Context, db *sql.DB) ([]byte, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("EXPLAIN /* pmm-agent */ %s", e.query))
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.Debug)
	w.Write([]byte(strings.Join(columns, "\t"))) //nolint:errcheck
	for rows.Next() {
		dest := make([]interface{}, len(columns))
		for i := range dest {
			var ns *nullString
			dest[i] = &ns
		}
		if err = rows.Scan(dest...); err != nil {
			return nil, err
		}

		row := "\n"
		for _, d := range dest {
			ns := *d.(**nullString)
			row += ns.String() + "\t"
		}
		w.Write([]byte(row)) //nolint:errcheck
	}

	if err = w.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *mysqlExplainAction) explainJSON(ctx context.Context, db *sql.DB) ([]byte, error) {
	var res string
	err := db.QueryRowContext(ctx, fmt.Sprintf("EXPLAIN /* pmm-agent */ FORMAT=JSON %s", e.query)).Scan(&res)
	return []byte(res), err
}
