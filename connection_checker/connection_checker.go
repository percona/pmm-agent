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

package connection_checker

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/percona/pmgo"
	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/api/inventorypb"
)

func New() *ConnectionChecker {
	return &ConnectionChecker{}
}

type ConnectionChecker struct {
}

func (c *ConnectionChecker) Check(msg *agentpb.CheckConnectionRequest) error {
	switch msg.Type {
	case inventorypb.ServiceType_MYSQL_SERVICE:
		return c.checkSQLConnection("mysql", msg.Dsn)
	case inventorypb.ServiceType_POSTGRESQL_SERVICE:
		return c.checkSQLConnection("postgres", msg.Dsn)
	case inventorypb.ServiceType_MONGODB_SERVICE:
		return c.checkMongoDBConnection(msg.Dsn)
	default:
		panic(fmt.Sprintf("unhandled service type: %v", msg.Type))
	}
	return nil
}

func (c *ConnectionChecker) checkMongoDBConnection(dsn string) error {
	dialer := pmgo.NewDialer()
	session, err := dialer.Dial(dsn)
	if err != nil {
		return err
	}
	defer session.Close()
	err = session.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (c *ConnectionChecker) checkSQLConnection(dbType string, dsn string) error {
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		return err
	}
	_, err = db.Exec(`SELECT 'pmm-agent'`)
	if err != nil {
		return err
	}
	return nil
}
