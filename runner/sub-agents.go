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
	"io"
	"os"
	"os/exec"
	"text/template"

	"github.com/percona/pmm/api/agent"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/utils/logger"
)

// SubAgent is structure for sub-agents.
type SubAgent struct {
	cmd    *exec.Cmd
	log    *logger.CircularWriter
	l      *logrus.Entry
	enable bool
	Params *agent.SetStateRequest_AgentProcess
	port   uint32

	runningChan chan struct{}
}

type templateParams struct {
	ListenPort uint32
}

// NewSubAgent creates new SubAgent.
func NewSubAgent(params *agent.SetStateRequest_AgentProcess, port uint32) *SubAgent {
	l := logrus.WithField("component", "runner").
		WithField("agentID", params.AgentId).
		WithField("type", params.Type)

	return &SubAgent{
		Params: params,
		log:    logger.New(10),
		l:      l,
		enable: false,
		port:   port,
	}
}

// Start starts sub-agent.
func (m *SubAgent) Start(ctx context.Context) error {
	if m.Running() {
		return fmt.Errorf("can't start the process, process is already running")
	}
	m.enable = true
	name := m.binary()
	args, err := m.args()
	if err != nil {
		m.l.Errorln(err)
		return err
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = m.Params.Env
	cmd.Stdout = io.MultiWriter(os.Stdout, m.log)
	cmd.Stderr = io.MultiWriter(os.Stderr, m.log)

	err = cmd.Start()
	if err != nil {
		return err
	}
	m.cmd = cmd
	m.runningChan = make(chan struct{})
	go func() {
		_ = m.cmd.Wait()
		if m.enable {
			m.runningChan <- struct{}{}
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
	if !m.Running() {
		return fmt.Errorf("can't kill the process, process is not running")
	}
	m.enable = false
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

// Running returns state of sub-agent.
func (m *SubAgent) Running() bool {
	return m.cmd != nil && m.cmd.ProcessState == nil
}

// Pid returns pid of running process
func (m *SubAgent) Pid() *int {
	if m.Running() {
		return &m.cmd.Process.Pid
	}
	return nil
}

// Port returns listen port
func (m *SubAgent) Port() uint32 {
	return m.port
}

func (m *SubAgent) args() ([]string, error) {
	params := templateParams{
		ListenPort: m.port,
	}
	var args []string
	for _, arg := range m.Params.Args {
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
	switch m.Params.Type {
	case agent.Type_MYSQLD_EXPORTER:
		return "mysqld_exporter"
	case agent.Type_NODE_EXPORTER:
		return "node_exporter"
	default:
		m.l.Panic("unhandled type of agent", m.Params.Type)
		return ""
	}
}
