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
	"net/url"
	"sync"

	"github.com/percona/pmm/api/agentlocalpb"
	"github.com/percona/pmm/api/agentpb"

	"github.com/percona/pmm-agent/config"
)

type agentsGetter interface {
	AgentsList() (res []*agentlocalpb.AgentInfo, err error)
}

// AgentLocalServer represents local agent api server.
type AgentLocalServer struct {
	cfg *config.Config

	rw             sync.RWMutex
	ag             agentsGetter
	serverMetadata *agentpb.AgentServerMetadata
}

// NewAgentLocalServer creates new local agent api server instance.
func NewAgentLocalServer(cfg *config.Config) *AgentLocalServer {
	return &AgentLocalServer{cfg: cfg}
}

// SetAgentsGetter sets new dependency which represents agentsGetter.
func (als *AgentLocalServer) SetAgentsGetter(ag agentsGetter) {
	als.rw.Lock()
	defer als.rw.Unlock()
	als.ag = ag
}

// SetMetadata sets new values of ServerMetadata.
func (als *AgentLocalServer) SetMetadata(md *agentpb.AgentServerMetadata) {
	als.rw.Lock()
	defer als.rw.Unlock()
	als.serverMetadata = md
}

func (als *AgentLocalServer) getMetadata() agentpb.AgentServerMetadata {
	als.rw.RLock()
	defer als.rw.RUnlock()
	return *als.serverMetadata
}

func (als *AgentLocalServer) getAgentsList() (res []*agentlocalpb.AgentInfo, err error) {
	als.rw.RLock()
	defer als.rw.RUnlock()

	if als.ag == nil {
		panic("agentsGetter is nil. Use SetAgentsGetter(ag agentsGetter) to set it.")
	}

	return als.ag.AgentsList()
}

// Status returns local agent status.
func (als *AgentLocalServer) Status(ctx context.Context, req *agentlocalpb.StatusRequest) (*agentlocalpb.StatusResponse, error) {
	md := als.getMetadata()

	var user *url.Userinfo
	switch {
	case als.cfg.Password != "":
		user = url.UserPassword(als.cfg.Username, als.cfg.Password)
	case als.cfg.Username != "":
		user = url.User(als.cfg.Username)
	}
	u := url.URL{
		Scheme: "https",
		User:   user,
		Host:   als.cfg.Address,
		Path:   "/",
	}
	srvInfo := &agentlocalpb.ServerInfo{
		Url:          u.String(),
		InsecureTls:  als.cfg.InsecureTLS,
		Version:      md.ServerVersion,
		LastPingTime: nil, // TODO: Add LastPingTime
		Latency:      nil, // TODO: Calculate and Add Latency
	}

	agentsInfo, err := als.getAgentsList()
	if err != nil {
		agentsInfo = []*agentlocalpb.AgentInfo{}
	}

	return &agentlocalpb.StatusResponse{
		AgentId:      als.cfg.ID,
		RunsOnNodeId: md.AgentRunsOnNodeID,
		ServerInfo:   srvInfo,
		AgentsInfo:   agentsInfo, // TODO: Add real AgentsInfo
	}, nil
}

// check interfaces
var (
	_ agentlocalpb.AgentLocalServer = (*AgentLocalServer)(nil)
)
