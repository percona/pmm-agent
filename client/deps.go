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

package client

import (
	"context"

	"github.com/percona/pmm/api/agentpb"
	spb "google.golang.org/genproto/googleapis/rpc/status"
)

//go:generate mockery -name=connectionChecker -case=snake -inpkg -testonly
//go:generate mockery -name=registry -case=snake -inpkg -testonly
//go:generate mockery -name=supervisor -case=snake -inpkg -testonly

// connectionChecker is a subset of methods of connectionchecker.ConnectionChecker used by this package.
// We use it instead of real type for testing and to avoid dependency cycle.
type connectionChecker interface {
	Check(ctx context.Context, req *agentpb.CheckConnectionRequest, id uint32) *agentpb.CheckConnectionResponse
}

// registry is a subset of methods of tunnels.Registry used by this package.
// We use it instead of real type for testing and to avoid dependency cycle.
type registry interface {
	SetState(tunnels map[string]*agentpb.SetStateRequest_Tunnel) map[string]*spb.Status
	Write(ctx context.Context, data *agentpb.TunnelData) *agentpb.TunnelDataAck
}

// supervisor is a subset of methods of supervisor.Supervisor used by this package.
// We use it instead of real type for testing and to avoid dependency cycle.
type supervisor interface {
	Changes() <-chan *agentpb.StateChangedRequest
	QANRequests() <-chan *agentpb.QANCollectRequest
	SetState(*agentpb.SetStateRequest)
}
