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

package tunnels

import (
	"context"
	"sort"
	"sync"

	"github.com/percona/pmm-agent/tunnels/conn"
	"github.com/percona/pmm/api/agentlocalpb"
	"github.com/percona/pmm/api/agentpb"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Registry struct {
	rw      sync.RWMutex
	tunnels map[string]tunnelInfo
}

type tunnelInfo struct {
	listenPort  uint16
	connectPort uint16
	connns      map[string]conn.Conn
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Run(ctx context.Context) {

}

func (r *Registry) TunnelsList() []*agentlocalpb.TunnelInfo {
	r.rw.RLock()
	defer r.rw.RUnlock()

	res := make([]*agentlocalpb.TunnelInfo, 0, len(r.tunnels))

	for id, tunnel := range r.tunnels {
		info := &agentlocalpb.TunnelInfo{
			TunnelId:           id,
			ListenPort:         uint32(tunnel.listenPort),
			ConnectPort:        uint32(tunnel.connectPort),
			CurrentConnections: 42, // TODO
		}
		res = append(res, info)
	}

	sort.Slice(res, func(i, j int) bool { return res[i].TunnelId < res[j].TunnelId })
	return res
}

func (r *Registry) SetState(tunnels map[string]*agentpb.SetStateRequest_Tunnel) map[string]*spb.Status {
	r.rw.Lock()
	defer r.rw.Unlock()

	res := make(map[string]*spb.Status, len(tunnels))

	// FIXME
	r.tunnels = map[string]tunnelInfo{}
	for id, t := range tunnels {
		if t.ListenPort != 0 && t.ConnectPort != 0 {
			res[id] = status.New(codes.InvalidArgument, "Both listen_port and connect_port are set.").Proto()
			continue
		}

		r.tunnels[id] = tunnelInfo{
			listenPort:  uint16(t.ListenPort),
			connectPort: uint16(t.ConnectPort),
		}
	}

	return res
}

func (r *Registry) Write(ctx context.Context, data *agentpb.TunnelData) *agentpb.TunnelDataAck {
	r.rw.RLock()
	t := r.tunnels[data.ConnectionId]

	r.rw.RUnlock()
	return nil
}
