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

package supervisor

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"

	"github.com/percona/pmm-agent/runner"
)

type Supervisor struct {
	rw     sync.RWMutex
	agents map[uint32]*runner.SubAgent
	l      *logrus.Entry
	ctx    context.Context
}

func NewSupervisor(ctx context.Context) *Supervisor {
	supervisor := &Supervisor{
		agents: make(map[uint32]*runner.SubAgent),
		l:      logrus.WithField("component", "supervisor"),
		ctx:    ctx,
	}
	return supervisor
}

func (s *Supervisor) Start(agentParams *runner.AgentParams) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	agent, ok := s.agents[agentParams.AgentId]
	if !ok {
		agent = runner.NewSubAgent(agentParams)
	}
	if agent.GetState() == runner.RUNNING {
		return fmt.Errorf("agent id=%d has already run", agentParams.AgentId)
	} else {
		err := agent.Start(s.ctx)
		if err != nil {
			return err
		}
		s.l.Debugf("agent %d is started", agentParams.AgentId)
		s.agents[agentParams.AgentId] = agent
		go s.watchSubAgent(agent)
		return nil
	}
}

func (s *Supervisor) Stop(id uint32) error {
	s.rw.Lock()
	defer s.rw.Unlock()

	agent, ok := s.agents[id]
	if !ok {
		return fmt.Errorf("agent with id %d not found", id)
	}
	err := agent.Stop()
	if err != nil {
		return err
	}
	delete(s.agents, id)
	return nil
}

func (s *Supervisor) watchSubAgent(agent *runner.SubAgent) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-agent.Done():
			err := agent.Restart(s.ctx)
			if err != nil {
				s.l.Warnf("Error on restarting agent %s", agent.String())
			}
		}
	}
}
