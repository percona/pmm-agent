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
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/percona/pmm/api/agentlocalpb"
	"github.com/percona/pmm/api/agentpb"
	"github.com/sirupsen/logrus"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/percona/pmm-agent/tunnels/conn"
)

type Registry struct {
	ctx context.Context
	l   *logrus.Entry

	rw      sync.RWMutex
	tunnels map[string]tunnelInfo
}

type tunnelInfo struct {
	listenPort  uint16
	connectPort uint16
	conns       map[string]*conn.Conn
	cancels     map[string]context.CancelFunc
}

func NewRegistry(ctx context.Context) *Registry {
	return &Registry{
		ctx: ctx,
		l:   logrus.WithField("component", "tunnels-registry"),
	}
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
			CurrentConnections: uint32(len(tunnel.conns)),
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
			conns:       make(map[string]*conn.Conn),
			cancels:     make(map[string]context.CancelFunc),
		}
	}

	return res
}

func (r *Registry) Write(ctx context.Context, data *agentpb.TunnelData) *agentpb.TunnelDataAck {
	r.rw.RLock()

	tunnel, ok := r.tunnels[data.TunnelId]
	if !ok {
		r.rw.RUnlock()
		return &agentpb.TunnelDataAck{
			Status: status.Newf(codes.NotFound, "Tunnel not configured: %s.", data.TunnelId).Proto(),
		}
	}

	connectionID := data.ConnectionId
	c := tunnel.conns[connectionID]

	r.rw.RUnlock()

	if c == nil {
		if tunnel.listenPort != 0 {
			panic("not implemented yet")
		}

		dialCtx, dialCancel := context.WithTimeout(ctx, 2*time.Second)
		defer dialCancel()
		d := net.Dialer{}
		netConn, err := d.DialContext(dialCtx, "tcp", fmt.Sprintf(":%d", tunnel.connectPort))
		if err != nil {
			return &agentpb.TunnelDataAck{
				Status: status.FromContextError(err).Proto(),
			}
		}

		c = conn.NewConn(netConn.(*net.TCPConn), r.l)
		connCtx, connCancel := context.WithCancel(r.ctx)
		go func() {
			c.Run(connCtx)

			r.rw.Lock()

			delete(r.tunnels[data.TunnelId].conns, connectionID)

			r.rw.Unlock()
		}()
		r.rw.Lock()
		r.tunnels[data.TunnelId].conns[connectionID] = c
		r.tunnels[data.TunnelId].cancels[connectionID] = connCancel
		r.rw.Unlock()
	}

	err := c.Write(data.Data)
	if data.Close {
		c.CloseWrite()
	}
	if err != nil {
		return &agentpb.TunnelDataAck{
			Status: status.FromContextError(err).Proto(),
		}
	}

	return nil
}
