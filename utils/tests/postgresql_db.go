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
	"net"
	"net/url"
	"regexp"
	"strconv"
	"testing"
	"time"

	_ "github.com/lib/pq" // register SQL driver
	"github.com/stretchr/testify/require"
)

// regexps to extract version numbers from the `SELECT version()` output
var (
	postgresDBRegexp  = regexp.MustCompile(`PostgreSQL ([\d\.]+)`)
	cockroachDBRegexp = regexp.MustCompile(`CockroachDB CCL (v[\d\.]+)`)
)

// GetTestPostgreSQLDSN returns DNS for PostgreSQL test database.
func GetTestPostgreSQLDSN(tb testing.TB) string {
	tb.Helper()

	if testing.Short() {
		tb.Skip("-short flag is passed, skipping test with real database.")
	}
	q := make(url.Values)
	q.Set("sslmode", "disable") // TODO: make it configurable

	u := &url.URL{
		Scheme:   "postgres",
		Host:     net.JoinHostPort("localhost", strconv.Itoa(int(15432))),
		Path:     "pmm-agent",
		User:     url.UserPassword("pmm-agent", "pmm-agent-password"),
		RawQuery: q.Encode(),
	}

	return u.String()
}

// OpenTestPostgreSQL opens connection to PostgreSQL test database.
func OpenTestPostgreSQL(tb testing.TB) *sql.DB {
	tb.Helper()

	db, err := sql.Open("postgres", GetTestPostgreSQLDSN(tb))
	if err == nil {
		db.SetMaxIdleConns(10)
		db.SetMaxOpenConns(10)
		db.SetConnMaxLifetime(0)

		// Wait until PostgreSQL is running up to 30 seconds.
		for i := 0; i < 30; i++ {
			if err = db.Ping(); err == nil {
				break
			}
			time.Sleep(time.Second)
		}
	}
	require.NoError(tb, err)
	return db
}

// PostgreSQLVendor represents PostgreSQL vendor (Oracle, Percona).
type PostgreSQLVendor string

// MySQL vendors.
const (
	PostgreSQL  PostgreSQLVendor = "PostgreSQL"
	CockroachDB PostgreSQLVendor = "CockroachDB"
)

// PostgreSQLVersion returns MAJOR.MINOR PostgreSQL version (e.g. "9.5", "10.0", etc.) and vendor.
func PostgreSQLVersion(tb testing.TB, db *sql.DB) (string, PostgreSQLVendor) {
	tb.Helper()
	var databaseVersion string

	err := db.QueryRow("SELECT version()").Scan(&databaseVersion)
	require.NoError(tb, err)

	engineVersion, engine := engineAndVersionFromPlainText(databaseVersion)
	//tb.Logf("version = %q (mm = %q), version_comment = %q (vendor = %q)", version, mm, comment, "postgres")
	return engineVersion, engine
}
func engineAndVersionFromPlainText(databaseVersion string) (string, PostgreSQLVendor) {
	var engine PostgreSQLVendor
	var engineVersion string
	switch {
	case postgresDBRegexp.MatchString(databaseVersion):
		engine = PostgreSQL
		submatch := postgresDBRegexp.FindStringSubmatch(databaseVersion)
		engineVersion = submatch[1]
	case cockroachDBRegexp.MatchString(databaseVersion):
		engine = CockroachDB
		submatch := cockroachDBRegexp.FindStringSubmatch(databaseVersion)
		engineVersion = submatch[1]
	}
	return engineVersion, engine
}
