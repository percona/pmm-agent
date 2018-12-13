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
	"github.com/percona/pmm/api/agent"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/config"
	"github.com/percona/pmm-agent/ports"
)

type Supervisor struct {
	rw     sync.Mutex
	agents map[uint32]*subAgent

	l        *logrus.Entry
	ctx      context.Context
	registry *ports.Registry
}

func NewSupervisor(ctx context.Context, portsCfg config.Ports) *Supervisor {
	supervisor := &Supervisor{
		agents:   make(map[uint32]*subAgent),
		l:        logrus.WithField("component", "supervisor"),
		ctx:      ctx,
		registry: ports.NewRegistry(portsCfg.Min, portsCfg.Max, nil),
	}
	return supervisor
}

// UpdateState starts or updates all agents placed in args and stops all agents not placed in args, but already run.
func (s *Supervisor) UpdateState(processes []*agent.SetStateRequest_AgentProcess) []*agent.SetStateResponse_AgentProcess {
	var agentProcessesStates []*agent.SetStateResponse_AgentProcess
	processesMaps := make(map[uint32]*agent.SetStateRequest_AgentProcess)
	var agentsToStart []uint32
	var agentsToRestart []uint32
	var agentsToStop []uint32

	s.rw.Lock()
	defer s.rw.Unlock()
	for _, agentProcess := range processes {
		subAgent, ok := s.agents[agentProcess.AgentId]
		if ok {
			if !proto.Equal(subAgent.params, agentProcess) {
				agentsToRestart = append(agentsToRestart, agentProcess.AgentId)
			} else {
				state := &agent.SetStateResponse_AgentProcess{
					AgentId:    agentProcess.AgentId,
					ListenPort: subAgent.port,
				}
				agentProcessesStates = append(agentProcessesStates, state)
			}
		} else {
			agentsToStart = append(agentsToStart, agentProcess.AgentId)
		}
		processesMaps[agentProcess.AgentId] = agentProcess
	}
	for id := range s.agents {
		if _, ok := processesMaps[id]; !ok {
			agentsToStop = append(agentsToStop, id)
		}
	}

	for _, id := range agentsToStop {
		port := s.agents[id].port
		err := s.stop(id)
		if err != nil {
			s.l.Error(err)
			continue
		}
		state := &agent.SetStateResponse_AgentProcess{
			AgentId:    id,
			ListenPort: port,
			Disabled:   true,
		}
		agentProcessesStates = append(agentProcessesStates, state)
	}

	for _, id := range agentsToStart {
		err := s.start(processesMaps[id])
		if err != nil {
			s.l.Error(err)
			continue
		}
		state := &agent.SetStateResponse_AgentProcess{
			AgentId:    id,
			ListenPort: s.agents[id].port,
		}
		agentProcessesStates = append(agentProcessesStates, state)
	}

	for _, id := range agentsToRestart {
		subAgent := s.agents[id]
		subAgent.params = processesMaps[id]
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
		go s.watchSubAgent(id, subAgent)
		state := &agent.SetStateResponse_AgentProcess{
			AgentId:    id,
			ListenPort: subAgent.port,
		}
		agentProcessesStates = append(agentProcessesStates, state)
	}

	return agentProcessesStates
}

func (s *Supervisor) start(agentParams *agent.SetStateRequest_AgentProcess) error {
	port, err := s.registry.Reserve()
	if err != nil {
		return err
	}

	subAgent, ok := s.agents[agentParams.AgentId]
	if !ok {
		subAgent = NewSubAgent(agentParams, port)
	}
	if subAgent.Running() {
		return fmt.Errorf("subAgent id=%d has already run", agentParams.AgentId)
	}

	err = subAgent.Start(s.ctx)
	if err != nil {
		_ = s.registry.Release(port)
		return err
	}
	s.l.Debugf("subAgent id=%d is started", agentParams.AgentId)
	s.agents[agentParams.AgentId] = subAgent
	go s.watchSubAgent(agentParams.AgentId, subAgent)
	return nil
}

func (s *Supervisor) stop(id uint32) error {
	subAgent, ok := s.agents[id]
	if !ok {
		return fmt.Errorf("subAgent with id %d not found", id)
	}
	err := subAgent.Stop()
	if err != nil {
		return err
	}
	_ = s.registry.Release(subAgent.port)
	delete(s.agents, id)
	return nil
}

func (s *Supervisor) watchSubAgent(id uint32, agent *subAgent) {
	restartCount := 0
	var startTime <-chan time.Time
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-agent.Disabled():
			return
		case <-agent.Restart():
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
