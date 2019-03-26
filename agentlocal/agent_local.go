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

	"github.com/percona/pmm/api/agentlocalpb"
)

type AgentLocalServer struct {
	AgentID      string
	RunsOnNodeID string
}

func (als *AgentLocalServer) Status(ctx context.Context, req *agentlocalpb.StatusRequest) (*agentlocalpb.StatusResponse, error) {
	return &agentlocalpb.StatusResponse{
		AgentId:      als.AgentID,
		RunsOnNodeId: als.RunsOnNodeID,
	}, nil
}

// check interfaces
var (
	_ agentlocalpb.AgentLocalServer = (*AgentLocalServer)(nil)
)
