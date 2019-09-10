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
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // register SQL driver
	"github.com/percona/pmm/api/agentpb"
)

type mysqlShowCreateTableAction struct {
	id     string
	params *agentpb.StartActionRequest_MySQLShowCreateTableParams
}

// NewMySQLShowCreateTableAction creates MySQL SHOW CREATE TABLE Action.
// This is an Action that can run `SHOW CREATE TABLE` command on MySQL service with given DSN.
func NewMySQLShowCreateTableAction(id string, params *agentpb.StartActionRequest_MySQLShowCreateTableParams) Action {
	return &mysqlShowCreateTableAction{
		id:     id,
		params: params,
	}
}

// ID returns an Action ID.
func (a *mysqlShowCreateTableAction) ID() string {
	return a.id
}

// Type returns an Action type.
func (a *mysqlShowCreateTableAction) Type() string {
	return "mysql-show-create-table"
}

// Run runs an Action and returns output and error.
func (a *mysqlShowCreateTableAction) Run(ctx context.Context) ([]byte, error) {
	// TODO Use sql.OpenDB with ctx when https://github.com/go-sql-driver/mysql/issues/671 is released
	// (likely in version 1.5.0).

	db, err := sql.Open("mysql", a.params.Dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close() //nolint:errcheck

	// use %#q to convert "table" to `"table"` and `table` to "`table`" to avoid SQL injections
	var tableName, tableDef string
	row := db.QueryRowContext(ctx, fmt.Sprintf("SHOW /* pmm-agent */ CREATE TABLE %#q", a.params.Table))
	if err = row.Scan(&tableName, &tableDef); err != nil {
		return nil, err
	}
	return []byte(tableDef), nil
}

func (a *mysqlShowCreateTableAction) sealed() {}
