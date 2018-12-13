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

package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/percona/pmm/api/agent"
	"github.com/stretchr/testify/assert"
)

func TestSubAgentArgs(t *testing.T) {
	type fields struct {
		params *agent.SetStateRequest_AgentProcess
		port   uint32
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			"No args test",
			fields{
				&agent.SetStateRequest_AgentProcess{
					Args: []string{},
				},
				1234,
			},
			nil,
			false,
		},
		{
			"Simple test",
			fields{
				&agent.SetStateRequest_AgentProcess{
					Args: []string{"-web.listen-address=127.0.0.1:{{ .ListenPort }}"},
				},
				1234,
			},
			[]string{"-web.listen-address=127.0.0.1:1234"},
			false,
		},
		{
			"Multiple args test",
			fields{
				&agent.SetStateRequest_AgentProcess{
					Args: []string{"-collect.binlog_size", "-web.listen-address=127.0.0.1:{{ .ListenPort }}"},
				},
				9175,
			},
			[]string{"-collect.binlog_size", "-web.listen-address=127.0.0.1:9175"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSubAgent(tt.fields.params, tt.fields.port)
			got, err := m.args()
			if (err != nil) != tt.wantErr {
				t.Errorf("subAgent.args() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRaceCondition(t *testing.T) {
	m := newSubAgent(&agent.SetStateRequest_AgentProcess{
		Type: agent.Type_MYSQLD_EXPORTER,
		Args: []string{"-web.listen-address=127.0.0.1:{{ .ListenPort }}"},
		Env: []string{
			`DATA_SOURCE_NAME="pmm:pmm@(127.0.0.1:3306)/pmm-managed-dev"`,
		},
	}, 12345)
	ctx, cancel := context.WithCancel(context.Background())
	err := m.Start(ctx)
	if err != nil {
		t.Errorf("subAgent.start() error = %v", err)
		cancel()
		return
	}
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	for {
		if !m.Running() {
			break
		}
	}
}
