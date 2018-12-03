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

package runner

import (
	"bytes"
	"context"
	"fmt"
	"github.com/percona/pmm/api/agent"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"text/template"

	"github.com/percona/pmm-agent/utils/logger"
)

type State int32

const (
	INVALID State = 0
	RUNNING State = 1
	STOPPED State = 2
	EXITED  State = 3
)

// AgentParams is params to run sub-agents.
type AgentParams struct {
	AgentId uint32
	Type    agent.Type
	Args    []string
	Env     []string
	Configs map[string]string
	Port    uint32
}

// SubAgent is structure for sub-agents.
type SubAgent struct {
	cmd    *exec.Cmd
	log    *logger.CircularWriter
	l      *logrus.Entry
	state  State
	params *AgentParams

	runningChan chan struct{}
}

type templateParams struct {
	ListenPort uint32
}

// NewSubAgent creates new SubAgent.
func NewSubAgent(params *AgentParams) *SubAgent {
	l := logrus.WithField("component", "runner").
		WithField("agentID", params.AgentId).
		WithField("type", params.Type)

	return &SubAgent{
		params: params,
		log:    logger.New(10),
		l:      l,
		state:  INVALID,
	}
}

// Start starts sub-agent.
func (m *SubAgent) Start(ctx context.Context) error {
	if m.GetState() == RUNNING {
		return fmt.Errorf("can't start the process, process is already running")
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
		m.state = EXITED
		return err
	}
	m.cmd = cmd
	m.state = RUNNING
	m.runningChan = make(chan struct{})
	go func() {
		_ = m.cmd.Wait()
		if m.state != STOPPED {
			m.state = EXITED
			close(m.runningChan)
		}
	}()
	return nil
}

// Done returns channel to restart agent.
func (m *SubAgent) Done() <-chan struct{} {
	r := m.runningChan
	return r
}

// Stop stops sub-agent
func (m *SubAgent) Stop() error {
	if m.GetState() != RUNNING {
		return fmt.Errorf("can't kill the process, process is not running")
	}
	m.state = STOPPED
	err := m.cmd.Process.Kill()
	if err != nil {
		return err
	}
	return nil
}

// GetLogs returns logs from sub-agent STDOut and STDErr.
func (m *SubAgent) GetLogs() []string {
	return m.log.Data()
}

// GetState returns state of sub-agent.
func (m *SubAgent) GetState() State {
	return m.state
}

func (m *SubAgent) args() ([]string, error) {
	params := templateParams{
		ListenPort: m.params.Port,
	}
	var args []string
	for _, arg := range m.params.Args {
		buffer := &bytes.Buffer{}
		tmpl, err := template.New(arg).Parse(arg)
		if err != nil {
			return nil, err
		}
		err = tmpl.Execute(buffer, params)
		if err != nil {
			return nil, err
		}
		args = append(args, buffer.String())
	}
	return args, nil
}

func (m *SubAgent) binary() string {
	switch m.params.Type {
	case agent.Type_MYSQLD_EXPORTER:
		return "mysqld_exporter"
	default:
		m.l.Panic("unhandled type of agent", m.params.Type)
		return ""
	}
}
