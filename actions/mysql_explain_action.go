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
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
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

type mysqlExplainAction struct {
	id     string
	dsn    string
	query  string
	format MysqlExplainOutputFormat

	db *sql.DB
}

func (p *mysqlExplainAction) sealed() {}

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

// ID returns unique Action id.
func (p *mysqlExplainAction) ID() string {
	return p.id
}

// Type returns Action type as as string.
func (p *mysqlExplainAction) Type() string {
	return "pt-mysql-explain"
}

// Run starts an Action. This method is blocking.
func (p *mysqlExplainAction) Run(ctx context.Context) ([]byte, error) {
	var err error
	p.db, err = sql.Open("mysql", p.dsn)
	if err != nil {
		return nil, err
	}
	defer p.db.Close()

	// Transaction because we need to ensure USE and EXPLAIN are run in one connection
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	cfg, err := mysql.ParseDSN(p.dsn)
	if err != nil {
		return nil, err
	}

	// If the query has a default db, use it; else, all tables need to be db-qualified
	// or EXPLAIN will throw an error.
	if cfg.DBName != "" {
		_, err = tx.ExecContext(ctx, fmt.Sprintf("USE %s", cfg.DBName))
		if err != nil {
			return nil, err
		}
	}

	var bytes []byte

	switch p.format {
	case ExplainFormatDefault:
		out, err := p.classicExplain(ctx, tx)
		if err != nil {
			return nil, err
		}

		bytes, err = json.Marshal(out)
		if err != nil {
			return nil, err
		}
	case ExplainFormatJSON:
		out, err := p.jsonExplain(ctx, tx)
		if err != nil {
			return nil, err
		}

		bytes, err = json.Marshal(out)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported output format")
	}

	return bytes, nil
}

type explainRow struct {
	Id           sql.NullInt64
	SelectType   sql.NullString
	Table        sql.NullString
	Partitions   sql.NullString // split by comma; since MySQL 5.1
	CreateTable  sql.NullString
	Type         sql.NullString
	PossibleKeys sql.NullString // split by comma
	Key          sql.NullString
	KeyLen       sql.NullString
	Ref          sql.NullString
	Rows         sql.NullInt64
	Filtered     sql.NullFloat64 // as of 5.7.3
	Extra        sql.NullString  // split by semicolon
}

func (p *mysqlExplainAction) classicExplain(ctx context.Context, tx *sql.Tx) ([]*explainRow, error) {
	// Partitions are introduced since MySQL 5.1
	// We can simply run EXPLAIN /*!50100 PARTITIONS*/ to get this column when it's available
	// without prior check for MySQL version.
	if strings.TrimSpace(p.query) == "" {
		return nil, errors.Errorf("cannot run EXPLAIN on an empty query example")
	}
	rows, err := tx.QueryContext(ctx, fmt.Sprintf("EXPLAIN %s", p.query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Go rows.Scan() expects exact number of columns
	// so when number of columns is undefined then the easiest way to
	// overcome this problem is to count received number of columns
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	nCols := len(columns)

	var out []*explainRow
	for rows.Next() {
		explainRow := &explainRow{}
		switch nCols {
		case 12: // MySQL 5.6 with "filtered"
			err = rows.Scan(
				&explainRow.Id,
				&explainRow.SelectType,
				&explainRow.Table,
				&explainRow.Partitions,
				&explainRow.Type,
				&explainRow.PossibleKeys,
				&explainRow.Key,
				&explainRow.KeyLen,
				&explainRow.Ref,
				&explainRow.Rows,
				&explainRow.Filtered, // here
				&explainRow.Extra,
			)
		default:
			err = errors.New("unsupported EXPLAIN format")
		}
		if err != nil {
			return nil, err
		}
		out = append(out, explainRow)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (p *mysqlExplainAction) jsonExplain(ctx context.Context, tx *sql.Tx) (string, error) {
	// EXPLAIN in JSON format is introduced since MySQL 5.6.5 and MariaDB 10.1.2
	// https://mariadb.com/kb/en/mariadb/explain-format-json/
	ok, err := p.versionConstraint(ctx, ">= 5.6.5, < 10.0.0 || >= 10.1.2")
	if !ok || err != nil {
		return "", err
	}

	explain := ""
	err = tx.QueryRowContext(ctx, fmt.Sprintf("EXPLAIN FORMAT=JSON %s", p.query)).Scan(&explain)
	if err != nil {
		return "", err
	}

	return explain, nil
}

// versionConstraint checks if version fits given constraint
func (p *mysqlExplainAction) versionConstraint(ctx context.Context, constraint string) (bool, error) {
	version, err := p.getGlobalVarString(ctx, "version")
	if err != nil {
		return false, err
	}

	// Strip everything after the first dash
	re := regexp.MustCompile("-.*$")
	version = re.ReplaceAllString(version, "")
	v, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}

	constraints, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, err
	}
	return constraints.Check(v), nil
}

func (p *mysqlExplainAction) getGlobalVarString(ctx context.Context, varName string) (string, error) {
	if err := p.db.Ping(); err != nil {
		return "", errors.New("not connected")
	}
	var value string
	err := p.db.QueryRowContext(ctx, "SELECT @@GLOBAL."+varName).Scan(&value)
	if val, ok := err.(*mysql.MySQLError); ok {
		if val.Number == 1193 /*ER_UNKNOWN_SYSTEM_VARIABLE*/ {
			return "", errors.New("unknown system variable")
		}
	}
	return value, nil
}
