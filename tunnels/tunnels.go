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

// Package tunnels provides support for tunnels.
package tunnels

// // Supervisor manages tunnels.
// type Supervisor struct {
// 	rw      sync.RWMutex
// 	tunnels map[string]*agentpb.SetStateRequest_Tunnel
// }

// func NewSupervisor() *Supervisor {
// 	return &Supervisor{}
// }

// func (s *Supervisor) Run(ctx context.Context) {

// }

// // TunnelsList returns info for all Tunnels managed by this supervisor.
// func (s *Supervisor) TunnelsList() []*agentlocalpb.TunnelInfo {
// 	// TODO
// 	return nil
// }

// // SetState starts or updates all agents placed in args and stops all agents not placed in args, but already run.
// func (s *Supervisor) SetState(tunnels map[string]*agentpb.SetStateRequest_Tunnel) {
// 	s.rw.Lock()
// 	defer s.rw.Unlock()

// 	s.tunnels = tunnels

// 	for id, t := range tunnels {
// 		if t.ListenPort != 0 {
// 			addr := fmt.Sprintf("127.0.0.1:%d", t.ListenPort)
// 			l, err := net.Listen("tcp", addr)
// 			if err != nil {
// 				panic(err)
// 			}

// 			for {
// 				conn, err := l.Accept()
// 				if err != nil {
// 					panic(err)
// 				}
// 			}
// 		}
// 	}
// }
