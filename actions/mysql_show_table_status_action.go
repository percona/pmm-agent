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

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql" // register SQL driver
	"github.com/percona/pmm/api/agentpb"
)

type mysqlShowTableStatusAction struct {
	id     string
	params *agentpb.StartActionRequest_MySQLShowTableStatusParams
}

// NewMySQLShowTableStatusAction creates MySQL SHOW TABLE STATUS Action.
// This is an Action that can run `SHOW TABLE STATUS` command on MySQL service with given DSN.
func NewMySQLShowTableStatusAction(id string, params *agentpb.StartActionRequest_MySQLShowTableStatusParams) Action {
	return &mysqlShowTableStatusAction{
		id:     id,
		params: params,
	}
}

// ID returns an Action ID.
func (e *mysqlShowTableStatusAction) ID() string {
	return e.id
}

// Type returns an Action type.
func (e *mysqlShowTableStatusAction) Type() string {
	return "mysql-table-status"
}

// Run runs an Action and returns output and error.
func (e *mysqlShowTableStatusAction) Run(ctx context.Context) ([]byte, error) {
	// TODO Use sql.OpenDB with ctx when https://github.com/go-sql-driver/mysql/issues/671 is released
	// (likely in version 1.5.0).

	db, err := sql.Open("mysql", e.params.Dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close() //nolint:errcheck

	cfg, err := mysql.ParseDSN(e.params.Dsn)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf("SHOW TABLE STATUS /* pmm-agent */ FROM %s WHERE Name='%s'", cfg.DBName, e.params.Table)) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	all := make([][]interface{}, 0)
	for rows.Next() {
		dest := make([]interface{}, len(columns))
		for i := range dest {
			var sp *string
			dest[i] = &sp
		}
		if err = rows.Scan(dest...); err != nil {
			return nil, err
		}
		all = append(all, dest)
	}

	data := make([]map[string]interface{}, 0)
	for _, r := range all {
		m := make(map[string]interface{})
		for i, c := range columns {
			m[c] = r[i]
		}
		data = append(data, m)
	}

	out, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (e *mysqlShowTableStatusAction) sealed() {}
