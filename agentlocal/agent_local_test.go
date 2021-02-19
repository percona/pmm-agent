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

package agentlocal

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentlocalpb"
	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/api/inventorypb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/pmm-agent/config"
)

func setup(t *testing.T) ([]*agentlocalpb.AgentInfo, []*agentlocalpb.TunnelInfo, *mockSupervisor, *mockRegistry, *mockClient, *config.Config) {
	t.Helper()

	agentInfo := []*agentlocalpb.AgentInfo{{
		AgentId:   "/agent_id/00000000-0000-4000-8000-000000000002",
		AgentType: inventorypb.AgentType_NODE_EXPORTER,
		Status:    inventorypb.AgentStatus_RUNNING,
	}}
	supervisor := new(mockSupervisor)
	supervisor.Test(t)
	supervisor.On("AgentsList").Return(agentInfo)
	t.Cleanup(func() { supervisor.AssertExpectations(t) })

	tunnelInfo := []*agentlocalpb.TunnelInfo{}
	registry := new(mockRegistry)
	registry.Test(t)
	registry.On("TunnelsList").Return(tunnelInfo)
	t.Cleanup(func() { registry.AssertExpectations(t) })

	client := new(mockClient)
	client.Test(t)
	client.On("GetServerConnectMetadata").Return(&agentpb.ServerConnectMetadata{
		AgentRunsOnNodeID: "/node_id/00000000-0000-4000-8000-000000000003",
		ServerVersion:     "2.0.0-dev",
	})
	t.Cleanup(func() { client.AssertExpectations(t) })

	cfg := &config.Config{
		ID: "/agent_id/00000000-0000-4000-8000-000000000001",
		Server: config.Server{
			Address:  "127.0.0.1:8443",
			Username: "username",
			Password: "password",
		},
	}

	return agentInfo, tunnelInfo, supervisor, registry, client, cfg
}

func TestServerStatus(t *testing.T) {
	t.Parallel()

	t.Run("without network info", func(t *testing.T) {
		t.Parallel()

		agentInfo, tunnelInfo, supervisor, registry, client, cfg := setup(t)
		s := NewServer(cfg, supervisor, registry, client, "/some/dir/pmm-agent.yaml")

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
			ConfigFilepath: "/some/dir/pmm-agent.yaml",
			TunnelsInfo:    tunnelInfo,
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("with network info", func(t *testing.T) {
		t.Parallel()

		agentInfo, tunnelInfo, supervisor, registry, client, cfg := setup(t)
		latency := 5 * time.Millisecond
		clockDrift := time.Second
		client.On("GetNetworkInformation").Return(latency, clockDrift, nil)
		s := NewServer(cfg, supervisor, registry, client, "/some/dir/pmm-agent.yaml")

		// with network info
		actual, err := s.Status(context.Background(), &agentlocalpb.StatusRequest{GetNetworkInfo: true})
		require.NoError(t, err)
		expected := &agentlocalpb.StatusResponse{
			AgentId:      "/agent_id/00000000-0000-4000-8000-000000000001",
			RunsOnNodeId: "/node_id/00000000-0000-4000-8000-000000000003",
			ServerInfo: &agentlocalpb.ServerInfo{
				Url:        "https://username:password@127.0.0.1:8443/",
				Version:    "2.0.0-dev",
				Latency:    ptypes.DurationProto(latency),
				ClockDrift: ptypes.DurationProto(clockDrift),
				Connected:  true,
			},
			AgentsInfo:     agentInfo,
			ConfigFilepath: "/some/dir/pmm-agent.yaml",
			TunnelsInfo:    tunnelInfo,
		}
		assert.Equal(t, expected, actual)
	})
}
