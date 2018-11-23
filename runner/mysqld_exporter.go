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
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"

	"github.com/percona/pmm-agent/utils/logger"
)

type MySQLdExporter struct {
	cmd    *exec.Cmd
	log    *logger.CircularWriter
	l      *logrus.Entry
	state  State
	params *AgentParams
}

func NewMySQLdExporter(params *AgentParams) *MySQLdExporter {
	return &MySQLdExporter{
		params: params,
		log:    logger.New(10),
		l:      logrus.WithField("component", "mysqld_exporter").WithField("AgentID", params.AgentId),
	}
}

func (m *MySQLdExporter) Start(ctx context.Context) error {
	if m.GetState() == RUNNING {
		return fmt.Errorf("can't start the process, process is already running")
	}
	name := "mysqld_exporter"
	args := append(m.params.Args, fmt.Sprintf("-web.listen-address=127.0.0.1:%d", m.params.Port))
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = m.params.Env
	cmd.Stdout = io.MultiWriter(os.Stdout, m.log)
	cmd.Stderr = io.MultiWriter(os.Stderr, m.log)

	err := cmd.Start()
	if err != nil {
		m.state = CRASHED
		return err
	}
	m.cmd = cmd
	m.state = RUNNING
	go cmd.Wait()
	return nil
}

func (m *MySQLdExporter) Stop() error {
	if m.GetState() != RUNNING {
		return fmt.Errorf("can't kill the process, process is not running")
	}
	err := m.cmd.Process.Kill()
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLdExporter) GetLogs() string {
	return m.log.String()
}

func (m *MySQLdExporter) GetState() State {
	if m.cmd == nil {
		m.state = INVALID
		return m.state
	}

	if !m.cmd.ProcessState.Exited() {
		m.state = RUNNING
	} else if m.cmd.ProcessState.Success() {
		m.state = STOPPED
	} else {
		m.state = CRASHED
	}
	return m.state
}
