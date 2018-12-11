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
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/percona/pmm-agent/ports"
	"github.com/percona/pmm/api/agent"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/runner"
)

type Supervisor struct {
	rw       sync.RWMutex
	agents   map[uint32]*runner.SubAgent
	l        *logrus.Entry
	ctx      context.Context
	registry *ports.Registry
}

func NewSupervisor(ctx context.Context) *Supervisor {
	supervisor := &Supervisor{
		agents:   make(map[uint32]*runner.SubAgent),
		l:        logrus.WithField("component", "supervisor"),
		ctx:      ctx,
		registry: ports.NewRegistry(10000, 20000, nil),
	}
	return supervisor
}

// UpdateState starts or updates all agents placed in args and stops all agents not placed in args, but already run.
func (s *Supervisor) UpdateState(processes []*agent.SetStateRequest_AgentProcess) []*agent.SetStateResponse_AgentProcess {
	var agentProcessesStates []*agent.SetStateResponse_AgentProcess
	processesMaps := make(map[uint32]bool)
	for _, agentProcess := range processes {
		var port uint32
		subAgent, ok := s.agents[agentProcess.AgentId]
		if ok {
			if !proto.Equal(subAgent.Params, agentProcess) {
				subAgent.Params = agentProcess
				err := subAgent.Stop()
				if err != nil {
					s.l.Error(err)
					continue
				}
				err = subAgent.Start(s.ctx)
				if err != nil {
					s.l.Error(err)
					continue
				}
			}
			port = subAgent.Port()
		} else {
			var err error
			port, err = s.registry.Reserve()
			if err != nil {
				s.l.Error(err)
				continue
			} else {
				err = s.Start(agentProcess, port)
				if err != nil {
					s.l.Error(err)
					_ = s.registry.Release(port)
					continue
				}
			}
		}
		processesMaps[agentProcess.AgentId] = true

		state := &agent.SetStateResponse_AgentProcess{
			AgentId:    agentProcess.AgentId,
			ListenPort: port,
			Disabled:   false,
		}
		agentProcessesStates = append(agentProcessesStates, state)
	}
	for id, subAgent := range s.agents {
		_, ok := processesMaps[id]
		if !ok {
			err := s.Stop(id)
			if err != nil {
				s.l.Error(err)
				continue
			}
			state := &agent.SetStateResponse_AgentProcess{
				AgentId:    id,
				ListenPort: subAgent.Port(),
				Disabled:   true,
			}
			agentProcessesStates = append(agentProcessesStates, state)
		}
	}

	return agentProcessesStates
}

// Start starts new sub-agent and adds it into map.
func (s *Supervisor) Start(agentParams *agent.SetStateRequest_AgentProcess, port uint32) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	subAgent, ok := s.agents[agentParams.AgentId]
	if !ok {
		subAgent = runner.NewSubAgent(agentParams, port)
	}
	if subAgent.Running() {
		return fmt.Errorf("subAgent id=%d has already run", agentParams.AgentId)
	} else {
		err := subAgent.Start(s.ctx)
		if err != nil {
			return err
		}
		s.l.Debugf("subAgent id=%d is started", agentParams.AgentId)
		s.agents[agentParams.AgentId] = subAgent
		go s.watchSubAgent(agentParams.AgentId, subAgent)
		return nil
	}
}

// Stop stops new sub-agent and adds it into map.
func (s *Supervisor) Stop(id uint32) error {
	s.rw.Lock()
	defer s.rw.Unlock()

	subAgent, ok := s.agents[id]
	if !ok {
		return fmt.Errorf("subAgent with id %d not found", id)
	}
	err := subAgent.Stop()
	if err != nil {
		return err
	}
	_ = s.registry.Release(subAgent.Port())
	delete(s.agents, id)
	return nil
}

func (s *Supervisor) watchSubAgent(id uint32, agent *runner.SubAgent) {
	restartCount := 0
	var startTime <-chan time.Time
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-agent.Done():
			max := math.Pow(2, float64(restartCount))
			delay := rand.Int63n(int64(max))
			startTime = time.After(time.Duration(delay) * time.Millisecond)
			s.l.Debugf("restarting agent in %d milliseconds", delay)
		case <-startTime:
			err := agent.Start(s.ctx)
			if err != nil {
				s.l.Warnf("Error on restarting agent with id %d", id)
			}
			restartCount++
		}
	}
}
