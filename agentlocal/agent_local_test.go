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

package agentlocal

import (
	"context"
	"github.com/golang/protobuf/ptypes/duration"
	"testing"
	"time"

	"github.com/percona/pmm-agent/common"

	"github.com/percona/pmm/api/agentlocalpb"
	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/api/inventorypb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/pmm-agent/config"
)

func TestServerStatus(t *testing.T) {
	setup := func (t *testing.T) ([]*agentlocalpb.AgentInfo, *mockSupervisor, *mockClient, *config.Config) {
		agentInfo := []*agentlocalpb.AgentInfo{{
			AgentId:   "/agent_id/00000000-0000-4000-8000-000000000002",
			AgentType: agentpb.Type_NODE_EXPORTER,
			Status:    inventorypb.AgentStatus_RUNNING,
			Logs:      nil,
		}}
		supervisor := new(mockSupervisor)
		supervisor.Test(t)
		supervisor.On("AgentsList").Return(agentInfo)
		client := new(mockClient)
		client.Test(t)
		client.On("GetAgentServerMetadata").Return(&agentpb.AgentServerMetadata{
			AgentRunsOnNodeID: "/node_id/00000000-0000-4000-8000-000000000003",
			ServerVersion:     "2.0.0-dev",
		})
		cfg := &config.Config{
			ID: "/agent_id/00000000-0000-4000-8000-000000000001",
			Server: config.Server{
				Address:  "127.0.0.1:8443",
				Username: "username",
				Password: "password",
			},
		}
		return agentInfo, supervisor, client, cfg
	}

	t.Run("without network info", func(t *testing.T) {
		agentInfo, supervisor, client, cfg := setup(t)
		defer supervisor.AssertExpectations(t)
		defer client.AssertExpectations(t)
		s := NewServer(cfg, supervisor, client, "/some/dir/pmm-agent.yaml")

		// without network info
		actual, err := s.Status(context.Background(), &agentlocalpb.StatusRequest{GetNetworkInfo: false})
		require.NoError(t, err)
		expected := &agentlocalpb.StatusResponse{
			AgentId:      "/agent_id/00000000-0000-4000-8000-000000000001",
			RunsOnNodeId: "/node_id/00000000-0000-4000-8000-000000000003",
			ServerInfo: &agentlocalpb.ServerInfo{
				Url:       "https://username:password@127.0.0.1:8443/",
				Version:   "2.0.0-dev",
				Connected: true,
			},
			AgentsInfo:     agentInfo,
			ConfigFilePath: "/some/dir/pmm-agent.yaml",
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("with network info", func(t *testing.T) {
		agentInfo, supervisor, client, cfg := setup(t)
		client.On("GetNetworkInformation").Return(&common.NetworkInformation{Ping: 5 * time.Millisecond, ClockDrift: 1 * time.Second}, nil)
		defer supervisor.AssertExpectations(t)
		defer client.AssertExpectations(t)
		s := NewServer(cfg, supervisor, client, "/some/dir/pmm-agent.yaml")

		// with network info
		actual, err := s.Status(context.Background(), &agentlocalpb.StatusRequest{GetNetworkInfo: true})
		require.NoError(t, err)
		expected := &agentlocalpb.StatusResponse{
			AgentId:      "/agent_id/00000000-0000-4000-8000-000000000001",
			RunsOnNodeId: "/node_id/00000000-0000-4000-8000-000000000003",
			ServerInfo: &agentlocalpb.ServerInfo{
				Url:        "https://username:password@127.0.0.1:8443/",
				Version:    "2.0.0-dev",
				ClockDrift: &duration.Duration{Seconds: 1},
				Latency:    &duration.Duration{Nanos: 5000000},
				Connected:  true,
			},
			AgentsInfo:     agentInfo,
			ConfigFilePath: "/some/dir/pmm-agent.yaml",
		}
		assert.Equal(t, expected, actual)
	})
}
