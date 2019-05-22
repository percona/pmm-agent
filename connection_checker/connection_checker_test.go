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
	"testing"

	"github.com/percona/pmm/api/inventorypb"

	"github.com/percona/pmm/api/agentpb"
)

func TestConnectionChecker_Check(t *testing.T) {
	tests := []struct {
		name    string
		msg     *agentpb.CheckConnectionRequest
		wantErr bool
	}{
		{
			name: "mysql",
			msg: &agentpb.CheckConnectionRequest{
				Dsn:  "root:root-password@tcp(127.0.0.1:3306)/?clientFoundRows=true&parseTime=true&timeout=1s",
				Type: inventorypb.ServiceType_MYSQL_SERVICE,
			},
			wantErr: false,
		},
		{
			name: "mysql wrong params",
			msg: &agentpb.CheckConnectionRequest{
				Dsn:  "pmm-agent:pmm-agent-wrong-password@tcp(127.0.0.1:3306)/?clientFoundRows=true&parseTime=true&timeout=1s",
				Type: inventorypb.ServiceType_MYSQL_SERVICE,
			},
			wantErr: true,
		},
		{
			name: "postgres",
			msg: &agentpb.CheckConnectionRequest{
				Dsn:  "postgres://pmm-agent:pmm-agent-password@127.0.0.1:5432/postgres?connect_timeout=1&sslmode=disable",
				Type: inventorypb.ServiceType_POSTGRESQL_SERVICE,
			},
			wantErr: false,
		},
		{
			name: "postgres wrong params",
			msg: &agentpb.CheckConnectionRequest{
				Dsn:  "postgres://pmm-agent:pmm-agent-wrong-password@127.0.0.1:5432/postgres?connect_timeout=1&sslmode=disable",
				Type: inventorypb.ServiceType_POSTGRESQL_SERVICE,
			},
			wantErr: true,
		},
		{
			name: "postgres",
			msg: &agentpb.CheckConnectionRequest{
				Dsn:  "mongodb://pmm-agent:root-password@127.0.0.1:27017/admin",
				Type: inventorypb.ServiceType_MONGODB_SERVICE,
			},
			wantErr: false,
		},
		{
			name: "postgres wrong params",
			msg: &agentpb.CheckConnectionRequest{
				Dsn:  "mongodb://pmm-agent:root-password-wrong@127.0.0.1:27017/admin",
				Type: inventorypb.ServiceType_MONGODB_SERVICE,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConnectionChecker{}
			if err := c.Check(tt.msg); (err != nil) != tt.wantErr {
				t.Errorf("ConnectionChecker.Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
