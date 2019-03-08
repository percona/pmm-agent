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

package tests

import (
	"database/sql"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

// OpenTestMySQL opens connection to MySQL test database.
func OpenTestMySQL(tb testing.TB) *sql.DB {
	tb.Helper()

	cfg := mysql.NewConfig()
	cfg.User = "root"
	cfg.Passwd = "root-password"
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"

	// required for reform
	cfg.ClientFoundRows = true
	cfg.ParseTime = true

	dsn := cfg.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err == nil {
		db.SetMaxIdleConns(10)
		db.SetMaxOpenConns(10)
		db.SetConnMaxLifetime(0)

		// Wait until MySQL is running up to 15 seconds.
		// Use Exec instead of Ping to fill performance_schema tables.
		for i := 0; i < 15; i++ {
			if _, err = db.Exec("SELECT 'test'"); err == nil {
				break
			}
			time.Sleep(time.Second)
		}
	}
	require.NoError(tb, err)
	return db
}
