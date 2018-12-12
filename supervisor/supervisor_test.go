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
	"syscall"
	"testing"
	"time"

	"github.com/percona/pmm/api/agent"
	"github.com/stretchr/testify/assert"
)

func agentProcessIsExists(t *testing.T, s *Supervisor, agentID uint32) (int, bool) {
	subAgent, ok := s.agents[agentID]
	if !ok {
		t.Errorf("Sub-agent not added to map")
		return 0, false
	}
	pid := *subAgent.Pid()
	procExists := processIsExists(pid)
	return pid, procExists
}

func processIsExists(pid int) bool {
	killErr := syscall.Kill(pid, syscall.Signal(0))
	procExists := killErr == nil
	return procExists
}

func checkResponse(t *testing.T, process *agent.SetStateResponse_AgentProcess, disabled bool) {
	expectedResponse := agent.SetStateResponse_AgentProcess{
		AgentId:    process.AgentId,
		ListenPort: process.AgentId + 9999,
		Disabled:   disabled,
	}
	assert.Equal(t, expectedResponse, *process)
}

func setup() (context.CancelFunc, *Supervisor, []string, []string) {
	ctx, cancel := context.WithCancel(context.TODO())
	s := NewSupervisor(ctx)
	arguments := []string{
		"-web.listen-address=127.0.0.1:{{ .ListenPort }}",
	}
	env := []string{
		`DATA_SOURCE_NAME="pmm:pmm@(127.0.0.1:3306)/pmm-managed-dev"`,
	}
	return cancel, s, arguments, env
}

func TestUpdateStateSimple(t *testing.T) {
	cancel, s, arguments, env := setup()
	defer cancel()

	var processes []*agent.SetStateRequest_AgentProcess
	agentsCount := uint32(5)
	for i := uint32(1); i <= agentsCount; i++ {
		processes = append(processes, &agent.SetStateRequest_AgentProcess{
			AgentId: i,
			Type:    agent.Type_MYSQLD_EXPORTER,
			Args:    arguments,
			Env:     env,
		})
	}

	response := s.UpdateState(processes)
	time.Sleep(1 * time.Second)
	for _, process := range response {
		checkResponse(t, process, false)
	}

	if uint32(len(s.agents)) != agentsCount || uint32(len(response)) != agentsCount {
		t.Errorf("%d agents started, expected %d", len(s.agents), agentsCount)
	}
	pids := make(map[uint32]int)
	for i, subAgent := range s.agents {
		pid := *subAgent.Pid()
		pids[i] = pid
		if !processIsExists(pid) {
			t.Errorf("Sub-agent with id %d is not run", pid)
		}
	}

	processes = []*agent.SetStateRequest_AgentProcess{}
	for i := uint32(4); i <= agentsCount; i++ {
		process := &agent.SetStateRequest_AgentProcess{
			AgentId: i,
			Type:    agent.Type_MYSQLD_EXPORTER,
			Args:    arguments,
			Env:     env,
		}
		processes = append(processes, process)
	}

	response = s.UpdateState(processes)
	if uint32(len(s.agents)) != 2 || uint32(len(response)) != agentsCount {
		t.Errorf("%d agents started, expected %d", len(s.agents), agentsCount)
	}

	for _, process := range response {
		checkResponse(t, process, process.AgentId < 4)
	}

	time.Sleep(1 * time.Second)

	for i := uint32(1); i <= agentsCount; i++ {
		procExists := processIsExists(pids[i])
		enabled := i >= 4
		if procExists != enabled {
			t.Errorf("Sub-agent pid %d is run = %v, expected %v", pids[i], procExists, enabled)
		}
	}
}

func TestSimpleStartStopSubAgent(t *testing.T) {
	cancel, s, arguments, env := setup()
	defer cancel()

	agentID := uint32(1)
	params := &agent.SetStateRequest_AgentProcess{
		AgentId: agentID,
		Type:    agent.Type_MYSQLD_EXPORTER,
		Args:    arguments,
		Env:     env,
	}
	err := s.Start(params, 10000)
	if err != nil {
		t.Errorf("Supervisor.Start() error = %v", err)
	}
	time.Sleep(1 * time.Second)
	pid, procExists := agentProcessIsExists(t, s, agentID)
	if !procExists {
		t.Errorf("Sub-agent process not found error = %v", err)
	}
	err = s.Stop(agentID)
	if err != nil {
		t.Errorf("Supervisor.Stop() error = %v", err)
	}
	time.Sleep(1 * time.Second)
	procExists = processIsExists(pid)
	if procExists {
		t.Errorf("sub-agent with pid %d is not stopped", pid)
	}
}

func TestContextDoneStopSubAgents(t *testing.T) {
	cancel, s, arguments, env := setup()

	params := &agent.SetStateRequest_AgentProcess{
		AgentId: 1,
		Type:    agent.Type_MYSQLD_EXPORTER,
		Args:    arguments,
		Env:     env,
	}
	err := s.Start(params, 10000)
	if err != nil {
		t.Errorf("Supervisor.Start() error = %v", err)
	}
	pid, procExists := agentProcessIsExists(t, s, 1)
	if !procExists {
		t.Errorf("Sub-agent process not found error = %v", err)
	}
	cancel()
	time.Sleep(1 * time.Second)
	procExists = processIsExists(pid)
	if procExists {
		t.Errorf("sub-agent with pid %d is not stopped", pid)
	}
}

func TestSupervisorStartTwice(t *testing.T) {
	cancel, s, arguments, env := setup()
	defer cancel()

	params := &agent.SetStateRequest_AgentProcess{
		AgentId: 1,
		Type:    agent.Type_MYSQLD_EXPORTER,
		Args:    arguments,
		Env:     env,
	}
	err := s.Start(params, 10000)
	if err != nil {
		t.Errorf("Supervisor.Start() error = %v", err)
	}
	time.Sleep(1 * time.Second)
	err = s.Start(params, 10001)
	if err == nil {
		t.Errorf("Starting sub-agent second time should return error")
	}
}
