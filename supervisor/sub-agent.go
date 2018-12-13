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
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"text/template"

	"github.com/percona/pmm/api/agent"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/utils/logger"
)

var closedChan = make(chan struct{})
var emptyChan = make(chan struct{}, 1)

func init() {
	close(closedChan)
}

// subAgent is structure for sub-agents.
type subAgent struct {
	log          *logger.CircularWriter
	l            *logrus.Entry
	params       *agent.SetStateRequest_AgentProcess
	port         uint32
	runningChan  chan struct{}
	disabledChan chan struct{}
	cmd          *exec.Cmd
}

type templateParams struct {
	ListenPort uint32
}

// NewSubAgent creates new subAgent.
func NewSubAgent(params *agent.SetStateRequest_AgentProcess, port uint32) *subAgent {
	l := logrus.WithField("component", "runner").
		WithField("agentID", params.AgentId).
		WithField("type", params.Type)

	return &subAgent{
		params:       params,
		log:          logger.New(10),
		l:            l,
		port:         port,
		runningChan:  closedChan,
		disabledChan: closedChan,
	}
}

// start starts sub-agent.
func (m *subAgent) Start(ctx context.Context) error {
	if m.Running() {
		return fmt.Errorf("can't start the process, process is already running")
	}
	m.disabledChan = make(chan struct{}, 1)
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
	m.runningChan = make(chan struct{}, 1)
	go func() {
		_ = m.cmd.Wait()
		close(m.runningChan)
	}()
	return nil
}

// Restart returns channel to restart agent.
func (m *subAgent) Restart() <-chan struct{} {
	select {
	case <-m.disabledChan:
		return emptyChan
	default:
		return m.runningChan
	}
}

// stop stops sub-agent
func (m *subAgent) Stop() error {
	if !m.Running() {
		return fmt.Errorf("can't kill the process, process is not running")
	}

	close(m.disabledChan)
	err := m.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		return err
	}
	return nil
}

// GetLogs returns logs from sub-agent STDOut and STDErr.
func (m *subAgent) GetLogs() []string {
	return m.log.Data()
}

// Running returns state of sub-agent.
func (m *subAgent) Running() bool {
	select {
	case <-m.runningChan:
		return false
	default:
		return true
	}
}

func (m *subAgent) pid() int {
	if m.Running() {
		return m.cmd.Process.Pid
	}
	return 0
}

func (m *subAgent) args() ([]string, error) {
	params := templateParams{
		ListenPort: m.port,
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

func (m *subAgent) Disabled() chan struct{} {
	return m.disabledChan
}
