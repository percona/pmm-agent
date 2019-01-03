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
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/percona/pmm/api/agent"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"github.com/percona/pmm-agent/config"
	"github.com/percona/pmm-agent/utils/logger"
)

// States.
// TODO Switch to agent.Status enum.
const (
	STARTING = "STARTING"
	RUNNING  = "RUNNING"
	WAITING  = "WAITING"
	STOPPING = "STOPPING"
	STOPPED  = "STOPPED"
)

// Agents for testing.
const (
	type_TESTING_NOT_FOUND = agent.Type(100500)
	type_TESTING_SLEEP     = agent.Type(100501)
)

// subAgent is structure for sub-agents.
type subAgent struct {
	ctx            context.Context
	paths          *config.Paths
	log            *logger.CircularWriter
	l              *logrus.Entry
	params         *agent.SetStateRequest_AgentProcess
	changesCh      chan string
	restartCounter *restartCounter

	wantStop chan struct{}

	cmd     *exec.Cmd
	cmdWait chan struct{}
}

func newSubAgent(ctx context.Context, paths *config.Paths, params *agent.SetStateRequest_AgentProcess) *subAgent {
	l := logrus.WithField("component", "sub-agent").
		WithField("agentID", params.AgentId).
		WithField("type", params.Type.String())

	changesCh := make(chan string, 10)

	sAgent := &subAgent{
		ctx:            ctx,
		paths:          paths,
		log:            logger.New(10),
		l:              l,
		params:         params,
		changesCh:      changesCh,
		restartCounter: &restartCounter{count: 1},
		wantStop:       make(chan struct{}),
	}

	go func() {
		<-ctx.Done()
		close(sAgent.wantStop)
	}()

	go sAgent.toStarting()

	return sAgent
}

// starting -> running;
// starting -> waiting;
func (sa *subAgent) toStarting() {
	sa.l.Infof("Starting...")
	sa.changesCh <- STARTING

	var name string
	switch sa.params.Type {
	case agent.Type_NODE_EXPORTER:
		name = sa.paths.NodeExporter
	case agent.Type_MYSQLD_EXPORTER:
		name = sa.paths.MySQLdExporter
	case type_TESTING_NOT_FOUND:
		name = "testing_not_found"
	case type_TESTING_SLEEP:
		name = "sleep"
	default:
		sa.l.Errorf("Failed to start: unhandled agent type %[1]s (%[1]d).", sa.params.Type)
		go sa.toWaiting()
		return
	}

	cmd := exec.Command(name, sa.params.Args...)
	cmd.Env = sa.params.Env
	cmd.Stdout = io.MultiWriter(os.Stdout, sa.log)
	cmd.Stderr = io.MultiWriter(os.Stderr, sa.log) // FIXME race
	// TODO a.cmd.SysProcAttr https://jira.percona.com/browse/PMM-3173
	cmdWait := make(chan struct{})

	sa.cmd = cmd
	sa.cmdWait = cmdWait

	if err := cmd.Start(); err != nil {
		sa.l.Errorf("Failed to start: %s.", err)
		go sa.toWaiting()
		return
	}
	go func() {
		cmd.Wait()
		close(cmdWait)
	}()

	select {
	case <-time.After(time.Second):
		go sa.toRunning()
	case <-cmdWait:
		sa.l.Errorf("Failed to start: %s.", sa.cmd.ProcessState)
		go sa.toWaiting()
	}
}

// running  -> stopping;
// running  -> waiting;
func (sa *subAgent) toRunning() {
	sa.l.Infof("Running.")
	sa.changesCh <- RUNNING

	sa.restartCounter.Reset()

	select {
	case <-sa.wantStop:
		go sa.toStopping()
	case <-sa.cmdWait:
		sa.l.Errorf("Exited: %s.", sa.cmd.ProcessState)
		go sa.toWaiting()
	}
}

// waiting  -> starting;
// waiting  -> stopped;
func (sa *subAgent) toWaiting() {
	sa.restartCounter.Inc()
	delay := sa.restartCounter.Delay()

	sa.l.Infof("Waiting %s.", delay)
	sa.changesCh <- WAITING

	select {
	case <-time.After(delay):
		go sa.toStarting()
	case <-sa.wantStop:
		go sa.toStopping()
	}
}

// stopping -> stopped;
func (sa *subAgent) toStopping() {
	sa.l.Infof("Stopping...")
	sa.changesCh <- STOPPING

	if sa.cmd.Process == nil {
		go sa.stopped()
		return
	}

	err := sa.cmd.Process.Signal(unix.SIGTERM)
	if err != nil {
		sa.l.Errorf("Failed to send SIGTERM: %s.", err)
	}

	select {
	case <-sa.cmdWait:
		// nothing
	case <-time.After(5 * time.Second):
		err := sa.cmd.Process.Signal(unix.SIGKILL)
		if err != nil {
			sa.l.Errorf("Failed to send SIGKILL: %s.", err)
		}
		<-sa.cmdWait
	}

	go sa.stopped()
}

func (sa *subAgent) stopped() {
	sa.l.Infof("Stopped.")
	sa.changesCh <- STOPPED

	close(sa.changesCh)
}

// Changes returns all state changes for current sub-agent.
func (sa *subAgent) Changes() <-chan string {
	return sa.changesCh
}

// GetLogs returns logs from sub-agent STDOut and STDErr.
func (sa *subAgent) GetLogs() []string {
	return sa.log.Data()
}
