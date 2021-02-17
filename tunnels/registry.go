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

	"github.com/percona/pmm/api/agentlocalpb"
	"github.com/percona/pmm/api/agentpb"
)

type Registry struct {
	rw             sync.RWMutex
	listenTunnels  map[string]tunnelInfo
	connectTunnels map[string]tunnelInfo
}

type tunnelInfo struct {
	port uint16
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Run(ctx context.Context) {

}

func (r *Registry) TunnelsList() []*agentlocalpb.TunnelInfo {
	r.rw.RLock()
	defer r.rw.RUnlock()

	res := make([]*agentlocalpb.TunnelInfo, 0, len(r.listenTunnels)+len(r.connectTunnels))

	for id, tunnel := range r.listenTunnels {
		info := &agentlocalpb.TunnelInfo{
			TunnelId:           id,
			ListenPort:         uint32(tunnel.port),
			CurrentConnections: 42, // TODO
		}
		res = append(res, info)
	}

	for id, tunnel := range r.connectTunnels {
		info := &agentlocalpb.TunnelInfo{
			TunnelId:           id,
			ConnectPort:        uint32(tunnel.port),
			CurrentConnections: 77, // TODO
		}
		res = append(res, info)
	}

	sort.Slice(res, func(i, j int) bool { return res[i].TunnelId < res[j].TunnelId })
	return res
}

func (r *Registry) SetState(tunnels map[string]*agentpb.SetStateRequest_Tunnel) {
	r.rw.Lock()
	defer r.rw.Unlock()

	// FIXME
	r.listenTunnels = map[string]tunnelInfo{}
	r.connectTunnels = map[string]tunnelInfo{}
	for id, t := range tunnels {
		if t.ListenPort != 0 {
			r.listenTunnels[id] = tunnelInfo{
				port: uint16(t.ListenPort),
			}
		} else {
			r.connectTunnels[id] = tunnelInfo{
				port: uint16(t.ConnectPort),
			}
		}
	}
}
