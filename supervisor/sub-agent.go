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
	"sync"
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

	cmdM    sync.Mutex
	cmd     *exec.Cmd
	cmdWait chan struct{}
}

func newSubAgent(ctx context.Context, paths *config.Paths, params *agent.SetStateRequest_AgentProcess) *subAgent {
	l := logrus.WithField("component", "sub-agent").
		WithField("agentID", params.AgentId).
		WithField("type", params.Type)

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
func (a *subAgent) toStarting() {
	a.l.Infof("Starting...")
	a.changesCh <- STARTING

	var name string
	switch a.params.Type {
	case agent.Type_NODE_EXPORTER:
		name = a.paths.NodeExporter
	case agent.Type_MYSQLD_EXPORTER:
		name = a.paths.MySQLdExporter
	case type_TESTING_NOT_FOUND:
		name = "testing_not_found"
	case type_TESTING_SLEEP:
		name = "sleep"
	default:
		a.l.Errorf("Failed to start: unhandled agent type %[1]s (%[1]d).", a.params.Type)
		go a.toWaiting()
		return
	}

	cmd := exec.Command(name, a.params.Args...)
	cmd.Env = a.params.Env
	cmd.Stdout = io.MultiWriter(os.Stdout, a.log)
	cmd.Stderr = io.MultiWriter(os.Stderr, a.log)
	cmdWait := make(chan struct{})

	a.cmdM.Lock()
	a.cmd = cmd
	a.cmdWait = cmdWait
	a.cmdM.Unlock()

	if err := cmd.Start(); err != nil {
		a.l.Errorf("Failed to start: %s.", err)
		go a.toWaiting()
		return
	}
	go func() {
		cmd.Wait()
		close(cmdWait)
	}()

	select {
	case <-time.After(time.Second):
		go a.toRunning()
	case <-cmdWait:
		a.l.Errorf("Failed to start: %s.", a.cmd.ProcessState)
		go a.toWaiting()
	}
}

// running  -> stopping;
// running  -> waiting;
func (a *subAgent) toRunning() {
	a.l.Infof("Running.")
	a.changesCh <- RUNNING

	a.restartCounter.Reset()

	select {
	case <-a.wantStop:
		go a.toStopping()
	case <-a.wait():
		a.l.Errorf("Exited: %s.", a.cmd.ProcessState)
		go a.toWaiting()
	}
}

// waiting  -> starting;
// waiting  -> stopped;
func (a *subAgent) toWaiting() {
	a.restartCounter.Inc()
	delay := a.restartCounter.Delay()

	a.l.Infof("Waiting %s.", delay)
	a.changesCh <- WAITING

	select {
	case <-time.After(delay):
		go a.toStarting()
	case <-a.wantStop:
		go a.toStopping()
	}
}

// stopping -> stopped;
func (a *subAgent) toStopping() {
	a.l.Infof("Stopping...")
	a.changesCh <- STOPPING

	if a.cmd.Process == nil {
		go a.stopped()
		return
	}

	err := a.cmd.Process.Signal(unix.SIGTERM)
	if err != nil {
		a.l.Errorf("Failed to send SIGTERM: %s.", err)
	}

	wait := a.wait()
	select {
	case <-wait:
		// nothing
	case <-time.After(5 * time.Second):
		err := a.cmd.Process.Signal(unix.SIGKILL)
		if err != nil {
			a.l.Errorf("Failed to send SIGKILL: %s.", err)
		}
		<-wait
	}

	go a.stopped()
}

func (a *subAgent) stopped() {
	a.l.Infof("Stopped.")
	a.changesCh <- STOPPED
	close(a.changesCh)
}

func (a *subAgent) wait() <-chan struct{} {
	a.cmdM.Lock()
	wait := a.cmdWait
	a.cmdM.Unlock()
	return wait
}

// Changes returns all state changes for current sub-agent.
func (a *subAgent) Changes() <-chan string {
	return a.changesCh
}

// GetLogs returns logs from sub-agent STDOut and STDErr.
func (a *subAgent) GetLogs() []string {
	return a.log.Data()
}
