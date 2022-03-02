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

package version

import (
	"context"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/mysql"
)

func TestGetMySQLVersion(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		fmt.Println("error creating mock database")
		return
	}
	defer sqlDB.Close() //nolint:errcheck

	t.Run("PerconaServer", func(t *testing.T) {
		columns := []string{"variable_name", "value"}
		mock.ExpectQuery("SHOW").
			WillReturnRows(sqlmock.NewRows(columns).AddRow("version", "8.0.26-17"))
		mock.ExpectQuery("SHOW").
			WillReturnRows(sqlmock.NewRows(columns).AddRow("version_comment", "Percona Server (GPL), Release 17, Revision d7119cd"))

		db := reform.NewDB(sqlDB, mysql.Dialect, reform.NewPrintfLogger(t.Logf))

		q := db.WithContext(context.Background())

		version, vendor, err := GetMySQLVersion(q)
		assert.Equal(t, "8.0", version)
		assert.Equal(t, "percona", vendor)
		assert.NoError(t, err)
	})

	t.Run("MySQL", func(t *testing.T) {
		columns := []string{"variable_name", "value"}
		mock.ExpectQuery("SHOW").
			WillReturnRows(sqlmock.NewRows(columns).AddRow("version", "8.0.28"))
		mock.ExpectQuery("SHOW").
			WillReturnRows(sqlmock.NewRows(columns).AddRow("version_comment", "MySQL Community Server - GPL"))

		db := reform.NewDB(sqlDB, mysql.Dialect, reform.NewPrintfLogger(t.Logf))

		q := db.WithContext(context.Background())

		version, vendor, err := GetMySQLVersion(q)
		assert.Equal(t, "8.0", version)
		assert.Equal(t, "oracle", vendor)
		assert.NoError(t, err)
	})

	t.Run("MariaDB", func(t *testing.T) {
		columns := []string{"variable_name", "value"}
		mock.ExpectQuery("SHOW").
			WillReturnRows(sqlmock.NewRows(columns).AddRow("version", "10.2.43-MariaDB-1:10.2.43+maria~bionic"))
		mock.ExpectQuery("SHOW").
			WillReturnRows(sqlmock.NewRows(columns).AddRow("version_comment", "mariadb.org binary distribution"))

		db := reform.NewDB(sqlDB, mysql.Dialect, reform.NewPrintfLogger(t.Logf))

		q := db.WithContext(context.Background())

		version, vendor, err := GetMySQLVersion(q)
		assert.Equal(t, "10.2", version)
		assert.Equal(t, "mariadb", vendor)
		assert.NoError(t, err)
	})
}
