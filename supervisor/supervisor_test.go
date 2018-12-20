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

	"github.com/sirupsen/logrus"

	"github.com/percona/pmm/api/agent"
	"github.com/stretchr/testify/assert"

	"github.com/percona/pmm-agent/config"
)

const sleepTime = 200 * time.Millisecond

func agentProcessIsExists(t *testing.T, s *Supervisor, agentID uint32) (int, bool) {
	subAgent, ok := s.agents[agentID]
	if !ok {
		t.Errorf("Sub-agent not added to map")
		return 0, false
	}
	pid := subAgent.pid()
	procExists := processIsExists(pid)
	return pid, procExists
}

func processIsExists(pid int) bool {
	killErr := syscall.Kill(pid, syscall.Signal(0))
	return killErr == nil
}

func checkResponse(t *testing.T, process *agent.SetStateResponse_AgentProcess, disabled bool) {
	expectedResponse := agent.SetStateResponse_AgentProcess{
		AgentId:    process.AgentId,
		ListenPort: process.AgentId + 10000,
		Disabled:   disabled,
	}
	assert.Equal(t, expectedResponse, *process)
}

func setup() (context.CancelFunc, *Supervisor, []string, []string) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx, cancel := context.WithCancel(context.TODO())
	s := NewSupervisor(ctx, config.Ports{Min: 10001, Max: 20000})
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
	for _, process := range response {
		checkResponse(t, process, false)
	}

	if uint32(len(s.agents)) != agentsCount || uint32(len(response)) != agentsCount {
		t.Errorf("%d agents started, expected %d", len(s.agents), agentsCount)
	}
	for _, subAgent := range s.agents {
		for {
			state := <-subAgent.Changes()
			if state == RUNNING {
				break
			}
		}
	}
	pids := make(map[uint32]int)
	for i, subAgent := range s.agents {
		pid := subAgent.pid()
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

	// Restart if params updated
	process := &agent.SetStateRequest_AgentProcess{
		AgentId: 3,
		Type:    agent.Type_MYSQLD_EXPORTER,
		Args:    append(arguments, "-collect.binlog_size"),
		Env:     env,
	}
	processes = append(processes, process)

	response = s.UpdateState(processes)
	if uint32(len(response)) != agentsCount {
		t.Errorf("%d process states returned, expected %d", len(response), agentsCount)
	}
	if uint32(len(s.agents)) != 3 {
		t.Errorf("%d agents works, expected %d", len(s.agents), 3)
	}

	for _, process := range response {
		checkResponse(t, process, process.AgentId < 3)
	}
	time.Sleep(sleepTime)

	assert.NotEqual(t, pids[3], s.agents[3].pid())
	for i := uint32(1); i <= agentsCount; i++ {
		procExists := processIsExists(pids[i])
		enabled := i >= 4
		if procExists != enabled {
			t.Errorf("Sub-agent pid %d is run = %v, expected %v", pids[i], procExists, enabled)
		}
	}
}

func TestSubAgentArgs(t *testing.T) {
	type fields struct {
		params *agent.SetStateRequest_AgentProcess
		port   uint32
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			"No args test",
			fields{
				&agent.SetStateRequest_AgentProcess{
					Args: []string{},
				},
				1234,
			},
			[]string{},
			false,
		},
		{
			"Simple test",
			fields{
				&agent.SetStateRequest_AgentProcess{
					Args: []string{"-web.listen-address=127.0.0.1:{{ .ListenPort }}"},
				},
				1234,
			},
			[]string{"-web.listen-address=127.0.0.1:1234"},
			false,
		},
		{
			"Multiple args test",
			fields{
				&agent.SetStateRequest_AgentProcess{
					Args: []string{"-collect.binlog_size", "-web.listen-address=127.0.0.1:{{ .ListenPort }}"},
				},
				9175,
			},
			[]string{"-collect.binlog_size", "-web.listen-address=127.0.0.1:9175"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			m := NewSupervisor(ctx, config.Ports{Min: 10000, Max: 20000})
			got, err := m.args(tt.fields.params.Args, templateParams{ListenPort: tt.fields.port})
			if (err != nil) != tt.wantErr {
				t.Errorf("subAgent.args() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSimpleStartStopSubAgent(t *testing.T) {
	cancel, s, arguments, env := setup()
	defer cancel()

	agentID := uint32(1)
	params := agent.SetStateRequest_AgentProcess{
		AgentId: agentID,
		Type:    agent.Type_MYSQLD_EXPORTER,
		Args:    arguments,
		Env:     env,
	}
	err := s.start(params, 12345)
	if err != nil {
		t.Errorf("Supervisor.start() error = %v", err)
	}
	for {
		state := <-s.agents[agentID].Changes()
		if state == RUNNING {
			break
		}
	}
	pid, procExists := agentProcessIsExists(t, s, agentID)
	if !procExists {
		t.Errorf("Sub-agent process not found error = %v", err)
	}
	s.stop(agentID, true)
	time.Sleep(sleepTime)
	procExists = processIsExists(pid)
	if procExists {
		t.Errorf("sub-agent with pid %d is not stopped", pid)
	}
}

func TestContextDoneStopSubAgents(t *testing.T) {
	cancel, s, arguments, env := setup()

	params := agent.SetStateRequest_AgentProcess{
		AgentId: 1,
		Type:    agent.Type_MYSQLD_EXPORTER,
		Args:    arguments,
		Env:     env,
	}
	err := s.start(params, 12346)
	if err != nil {
		t.Errorf("Supervisor.start() error = %v", err)
	}
	for {
		state := <-s.agents[1].Changes()
		if state == RUNNING {
			break
		}
	}
	pid, procExists := agentProcessIsExists(t, s, 1)
	if !procExists {
		t.Errorf("Sub-agent process not found error = %v", err)
	}
	cancel()
	time.Sleep(sleepTime)
	procExists = processIsExists(pid)
	if procExists {
		t.Errorf("sub-agent with pid %d is not stopped", pid)
	}
}
