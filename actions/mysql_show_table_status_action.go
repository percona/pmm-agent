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

	status := &showTableStatus{}
	err = db.QueryRowContext(ctx, fmt.Sprintf("SHOW TABLE STATUS FROM %s WHERE Name='%s'", cfg.DBName, e.params.Table)).Scan(
		&status.Name,
		&status.Engine,
		&status.Version,
		&status.RowFormat,
		&status.Rows,
		&status.AvgRowLength,
		&status.DataLength,
		&status.MaxDataLength,
		&status.IndexLength,
		&status.DataFree,
		&status.AutoIncrement,
		&status.CreateTime,
		&status.UpdateTime,
		&status.CheckTime,
		&status.Collation,
		&status.Checksum,
		&status.CreateOptions,
		&status.Comment,
	)
	if err != nil {
		return nil, err
	}

	out, err := json.Marshal(status)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (e *mysqlShowTableStatusAction) sealed() {}

type showTableStatus struct {
	Name          string         `json:"name"`
	Engine        string         `json:"engine"`
	Version       string         `json:"version"`
	RowFormat     string         `json:"row_format"`
	Rows          sql.NullInt64  `json:"rows"`
	AvgRowLength  sql.NullInt64  `json:"avg_row_length"`
	DataLength    sql.NullInt64  `json:"data_length"`
	MaxDataLength sql.NullInt64  `json:"max_data_length"`
	IndexLength   sql.NullInt64  `json:"index_length"`
	DataFree      sql.NullInt64  `json:"data_free"`
	AutoIncrement sql.NullInt64  `json:"auto_increment"`
	CreateTime    mysql.NullTime `json:"create_time"`
	UpdateTime    mysql.NullTime `json:"update_time"`
	CheckTime     mysql.NullTime `json:"check_time"`
	Collation     sql.NullString `json:"collation"`
	Checksum      sql.NullString `json:"checksum"`
	CreateOptions sql.NullString `json:"create_options"`
	Comment       sql.NullString `json:"comment"`
}
