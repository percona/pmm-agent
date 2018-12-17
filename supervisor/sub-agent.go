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
	"io"
	"os"
	"os/exec"
	"syscall"
	"text/template"

	"github.com/looplab/fsm"
	"github.com/percona/pmm/api/agent"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/utils/logger"
)

const (
	NEW      = "new"
	STARTING = "starting"
	RUNNING  = "running"
	EXITED   = "exited"
	BACKOFF  = "backoff"
	STOPPING = "stopping"
	STOPPED  = "stopped"
)

// subAgent is structure for sub-agents.
type subAgent struct {
	log    *logger.CircularWriter
	l      *logrus.Entry
	params *agent.SetStateRequest_AgentProcess
	port   uint32
	cmd    *exec.Cmd
	state  *fsm.FSM
}

type templateParams struct {
	ListenPort uint32
}

func newSubAgent(params *agent.SetStateRequest_AgentProcess, port uint32) *subAgent {
	l := logrus.WithField("component", "runner").
		WithField("agentID", params.AgentId).
		WithField("type", params.Type)
	var closedChan = make(chan struct{})
	close(closedChan)

	state := fsm.NewFSM(
		NEW,
		fsm.Events{
			{Name: STARTING, Src: []string{NEW, BACKOFF, STOPPED}, Dst: STARTING},
			{Name: RUNNING, Src: []string{STARTING}, Dst: RUNNING},
			{Name: EXITED, Src: []string{RUNNING}, Dst: EXITED},
			{Name: BACKOFF, Src: []string{EXITED, STARTING}, Dst: BACKOFF},
			{Name: STOPPING, Src: []string{RUNNING}, Dst: STOPPING},
			{Name: STOPPED, Src: []string{STOPPING, BACKOFF}, Dst: STOPPED},
		},
		fsm.Callbacks{},
	)

	return &subAgent{
		params: params,
		log:    logger.New(10),
		l:      l,
		port:   port,
		state:  state,
	}
}

// start starts sub-agent.
func (m *subAgent) Start(ctx context.Context) error {
	err := m.state.Event(STARTING)
	if err != nil {
		return err
	}
	name := m.binary()
	args, err := m.args()
	if err != nil {
		m.l.Errorln(err)
		return err
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = m.params.Env
	cmd.Stdout = io.MultiWriter(os.Stdout, m.log)
	cmd.Stderr = io.MultiWriter(os.Stderr, m.log)

	err = cmd.Start()
	if err != nil {
		return err
	}
	m.cmd = cmd
	err = m.state.Event(RUNNING)
	if err != nil {
		return err
	}
	go func() {
		err := m.cmd.Wait()
		if err != nil {
			m.l.Debug(err)
		}
		err = m.state.Event(EXITED)
		if err != nil {
			m.l.Debug(err)
		}
	}()
	return nil
}

// stop stops sub-agent
func (m *subAgent) Stop() error {
	err := m.state.Event(STOPPING)
	if err != nil {
		return err
	}

	err = m.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		return err
	}

	if syscall.Kill(m.cmd.Process.Pid, syscall.Signal(0)) == nil {
		err := m.cmd.Process.Kill()
		if err != nil {
			return err
		}
	}

	err = m.state.Event(STOPPED)
	if err != nil {
		return err
	}

	return nil
}

// GetLogs returns logs from sub-agent STDOut and STDErr.
func (m *subAgent) GetLogs() []string {
	return m.log.Data()
}

// GetState returns state of sub-agent
func (m *subAgent) GetState() string {
	return m.state.Current()
}

func (m *subAgent) pid() int {
	if m.state.Is(RUNNING) {
		return m.cmd.Process.Pid
	}
	return 0
}

func (m *subAgent) args() ([]string, error) {
	params := templateParams{
		ListenPort: m.port,
	}
	args := make([]string, len(m.params.Args))
	for i, arg := range m.params.Args {
		buffer := &bytes.Buffer{}
		tmpl, err := template.New(arg).Parse(arg)
		if err != nil {
			return nil, err
		}
		err = tmpl.Execute(buffer, params)
		if err != nil {
			return nil, err
		}
		args[i] = buffer.String()
	}
	return args, nil
}

func (m *subAgent) binary() string {
	switch m.params.Type {
	case agent.Type_MYSQLD_EXPORTER:
		return "mysqld_exporter"
	case agent.Type_NODE_EXPORTER:
		return "node_exporter"
	default:
		m.l.Panic("unhandled type of agent", m.params.Type)
		return ""
	}
}
