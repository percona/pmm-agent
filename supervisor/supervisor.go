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
	"bytes"
	"context"
	"sync"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/percona/pmm/api/agent"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/config"
	"github.com/percona/pmm-agent/ports"
)

// StateUpdate contains info about agent and current state.
type StateUpdate struct {
	AgentID uint32
	State   string
}

type subAgentInfo struct {
	*subAgent
	cancel func()
}

// Supervisor manages all agents.
type Supervisor struct {
	rw     sync.Mutex
	agents map[uint32]*subAgentInfo
	params map[uint32]*agent.SetStateRequest_AgentProcess
	ports  map[uint32]uint32

	l        *logrus.Entry
	ctx      context.Context
	registry *ports.Registry
	changes  chan StateUpdate
}

type templateParams struct {
	ListenPort uint32
}

// NewSupervisor creates new Supervisor object.
func NewSupervisor(ctx context.Context, portsCfg config.Ports) *Supervisor {
	supervisor := &Supervisor{
		agents:   make(map[uint32]*subAgentInfo),
		params:   make(map[uint32]*agent.SetStateRequest_AgentProcess),
		ports:    make(map[uint32]uint32),
		l:        logrus.WithField("component", "supervisor"),
		ctx:      ctx,
		registry: ports.NewRegistry(portsCfg.Min, portsCfg.Max, nil),
		changes:  make(chan StateUpdate),
	}
	go supervisor.handleContextDone()
	return supervisor
}

// UpdateState starts or updates all agents placed in args and stops all agents not placed in args, but already run.
func (s *Supervisor) UpdateState(processes []*agent.SetStateRequest_AgentProcess) []*agent.SetStateResponse_AgentProcess {
	var agentProcessesStates []*agent.SetStateResponse_AgentProcess
	processesMaps := make(map[uint32]*agent.SetStateRequest_AgentProcess)
	var toStart, toRestart, toStop []uint32

	s.rw.Lock()
	defer s.rw.Unlock()
	for _, agentProcess := range processes {
		_, ok := s.agents[agentProcess.AgentId]
		switch {
		case !ok:
			toStart = append(toStart, agentProcess.AgentId)
		case ok && !proto.Equal(s.params[agentProcess.AgentId], agentProcess):
			toRestart = append(toRestart, agentProcess.AgentId)
		default:
			state := &agent.SetStateResponse_AgentProcess{
				AgentId:    agentProcess.AgentId,
				ListenPort: s.ports[agentProcess.AgentId],
			}
			agentProcessesStates = append(agentProcessesStates, state)
		}
		processesMaps[agentProcess.AgentId] = agentProcess
	}
	for id := range s.agents {
		if _, ok := processesMaps[id]; !ok {
			toStop = append(toStop, id)
		}
	}

	for _, id := range toStart {
		port, err := s.registry.Reserve()
		if err != nil {
			continue
		}

		err = s.start(*processesMaps[id], port)
		if err != nil {
			s.l.Error(err)
			continue
		}
		s.params[id] = processesMaps[id]
		state := &agent.SetStateResponse_AgentProcess{
			AgentId:    id,
			ListenPort: port,
		}
		agentProcessesStates = append(agentProcessesStates, state)
	}

	for _, id := range toStop {
		port := s.ports[id]
		s.stop(id, false)
		state := &agent.SetStateResponse_AgentProcess{
			AgentId:    id,
			ListenPort: port,
			Disabled:   true,
		}
		agentProcessesStates = append(agentProcessesStates, state)
	}

	for _, id := range toRestart {
		port := s.ports[id]
		s.stop(id, true)
		err := s.start(*processesMaps[id], port)
		if err != nil {
			s.l.Error(err)
			continue
		}
		s.params[id] = processesMaps[id]
		state := &agent.SetStateResponse_AgentProcess{
			AgentId:    id,
			ListenPort: port,
		}
		agentProcessesStates = append(agentProcessesStates, state)
	}

	return agentProcessesStates
}

// StateUpdates returns update changes for all agents
func (s *Supervisor) StateUpdates() <-chan StateUpdate {
	return s.changes
}

func (s *Supervisor) start(agentParams agent.SetStateRequest_AgentProcess, port uint32) (err error) {
	agentParams.Args, err = s.args(agentParams.Args, templateParams{ListenPort: port})
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(s.ctx)
	subAgent := newSubAgent(ctx, &agentParams)
	go s.watchUpdates(agentParams.AgentId, subAgent)

	s.l.Debugf("subAgent id=%d is started", agentParams.AgentId)
	s.agents[agentParams.AgentId] = &subAgentInfo{
		subAgent: subAgent,
		cancel:   cancel,
	}
	s.ports[agentParams.AgentId] = port
	return
}

func (s *Supervisor) stop(id uint32, wait bool) {
	subAgentInfo := s.agents[id]
	subAgentInfo.cancel()
	if wait {
		for {
			_, more := <-subAgentInfo.Changes()
			if !more {
				break
			}
		}
	}

	_ = s.registry.Release(s.ports[id])
	delete(s.agents, id)
	delete(s.ports, id)
	delete(s.params, id)
}

func (s *Supervisor) args(args []string, params templateParams) ([]string, error) {
	result := make([]string, len(args))
	for i, arg := range args {
		buffer := &bytes.Buffer{}
		tmpl, err := template.New(arg).Parse(arg)
		if err != nil {
			return nil, err
		}
		err = tmpl.Execute(buffer, params)
		if err != nil {
			return nil, err
		}
		result[i] = buffer.String()
	}
	return result, nil
}

func (s *Supervisor) watchUpdates(id uint32, sa *subAgent) {
	for {
		select {
		case state, more := <-sa.Changes():
			s.changes <- StateUpdate{AgentID: id, State: state}
			if !more {
				return
			}
		}
	}
}

func (s *Supervisor) handleContextDone() {

	for {
		select {
		case <-s.ctx.Done():
			s.rw.Lock()
			for i := range s.agents {
				s.stop(i, false)
			}
			s.rw.Unlock()
			return
		}
	}
}
