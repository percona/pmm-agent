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

package qan

import (
	"context"

	"github.com/percona/pmm/api/agent"
)

// Supervisor manages all internal agents.
type Supervisor struct {
}

// NewSupervisor creates new Supervisor object.
func NewSupervisor(ctx context.Context) *Supervisor {
	return &Supervisor{}
}

func (s *Supervisor) SetState(internalAgents map[string]*agent.SetStateRequest_InternalAgent) {
}

// Changes returns channel with agent's state changes.
func (s *Supervisor) Changes() <-chan agent.StateChangedRequest {
	return nil
}

func (s *Supervisor) Data() <-chan *agent.QANData {
	return nil
}
