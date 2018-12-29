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
	"syscall"
	"time"

	"github.com/looplab/fsm"
	"github.com/percona/pmm/api/agent"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/utils/logger"
)

//States
const (
	NEW      = "new"
	STARTING = "starting"
	RUNNING  = "running"
	BACKOFF  = "backoff"
	STOPPING = "stopping"
	STOPPED  = "stopped"
)

//Events
const (
	START   = "start"
	STARTED = "started"
	RESTART = "restart"
	STOP    = "stop"
	EXIT    = "exit"
)

// subAgent is structure for sub-agents.
type subAgent struct {
	log            *logger.CircularWriter
	l              *logrus.Entry
	params         *agent.SetStateRequest_AgentProcess
	port           uint32
	state          *fsm.FSM
	changesChan    chan string
	restartCounter *restartCounter

	md       sync.Mutex
	doneChan chan struct{}

	mc  sync.Mutex
	cmd *exec.Cmd
}

func newSubAgent(ctx context.Context, params *agent.SetStateRequest_AgentProcess) *subAgent {
	l := logrus.WithField("component", "sub-agent").
		WithField("agentID", params.AgentId).
		WithField("type", params.Type)

	changesChan := make(chan string, 10)

	state := fsm.NewFSM(
		NEW,
		fsm.Events{
			{Name: START, Src: []string{NEW, BACKOFF}, Dst: STARTING},
			{Name: STARTED, Src: []string{STARTING}, Dst: RUNNING},
			{Name: RESTART, Src: []string{STARTING, RUNNING}, Dst: BACKOFF},
			{Name: STOP, Src: []string{RUNNING, BACKOFF}, Dst: STOPPING},
			{Name: EXIT, Src: []string{STOPPING, BACKOFF}, Dst: STOPPED},
		},
		fsm.Callbacks{
			"enter_state": func(event *fsm.Event) {
				changesChan <- event.Dst
			},
			"after_exit": func(event *fsm.Event) {
				close(changesChan)
			},
		},
	)

	sAgent := &subAgent{
		params:         params,
		log:            logger.New(10),
		l:              l,
		state:          state,
		changesChan:    changesChan,
		restartCounter: &restartCounter{count: 1},
		doneChan:       make(chan struct{}),
	}
	go sAgent.start(ctx)
	return sAgent
}

// start starts sub-agent.
func (a *subAgent) start(ctx context.Context) {
	filename := a.binary()
	for a.state.Can(START) {

		a.mc.Lock()
		err := a.state.Event(START)
		if err != nil {
			a.l.Debug(err)
			a.mc.Unlock()
			return
		}
		a.md.Lock()
		a.doneChan = make(chan struct{})
		a.md.Unlock()
		cmd := exec.Command(filename, a.params.Args...)
		cmd.Env = a.params.Env
		cmd.Stdout = io.MultiWriter(os.Stdout, a.log)
		cmd.Stderr = io.MultiWriter(os.Stderr, a.log)
		err = cmd.Start()
		if err != nil {
			a.l.Debug(err)
			a.mc.Unlock()
			if !a.backoff(ctx) {
				return
			}
		}
		a.cmd = cmd
		err = a.state.Event(STARTED)
		a.mc.Unlock()
		if err != nil {
			a.l.Debug(err)
			if !a.backoff(ctx) {
				return
			}
		}
		go func() {
			resetTime := time.After(2 * time.Second)
			select {
			case <-resetTime:
				a.restartCounter.Reset()
			case <-a.done():
				return
			}
		}()
		err = a.cmd.Wait()
		if err != nil {
			a.l.Debugf("sub-agent exited with message: %s", err)
		}
		close(a.doneChan)
		err = a.state.Event(RESTART)
		if err != nil {
			a.l.Debug(err)
			return
		}
		if !a.backoff(ctx) {
			return
		}
	}
}

func (a *subAgent) backoff(ctx context.Context) bool {
	delay := a.restartCounter.Delay()
	startTime := time.After(delay)
	for {
		select {
		case <-ctx.Done():
			err := a.state.Event(EXIT)
			if err != nil {
				a.l.Debug(err)
			}
			a.l.Debugf("exited on context done")
			return false
		case <-startTime:
			a.l.Debugf("restarted after %v", delay)
			a.restartCounter.Inc()
			return true
		}
	}
}

// stop stops sub-agent
func (a *subAgent) Stop() {
	go func() {
		a.mc.Lock()
		defer a.mc.Unlock()
		err := a.state.Event(STOP)
		if err != nil {
			a.l.Warnln("Can't change state to STOPPING", err)
			return
		}

		err = a.cmd.Process.Signal(syscall.SIGINT)
		if err != nil {
			a.l.Warnln("Can't stop sub-agent", err)
			return
		}

		killTime := time.After(5 * time.Second)

		select {
		case <-killTime:
			err := a.cmd.Process.Kill()
			if err != nil {
				a.l.Warnln("Can't kill sub-agent", err)
				return
			}
		case <-a.done():
			break
		}

		err = a.state.Event(EXIT)
		if err != nil {
			a.l.Warnln("Can't change state to STOPPING", err)
			return
		}
	}()
}

func (a *subAgent) done() <-chan struct{} {
	a.md.Lock()
	done := a.doneChan
	a.md.Unlock()
	return done
}

// GetLogs returns logs from sub-agent STDOut and STDErr.
func (a *subAgent) GetLogs() []string {
	return a.log.Data()
}

func (a *subAgent) Changes() <-chan string {
	return a.changesChan
}

func (a *subAgent) binary() string {
	switch a.params.Type {
	case agent.Type_MYSQLD_EXPORTER:
		return "mysqld_exporter"
	case agent.Type_NODE_EXPORTER:
		return "node_exporter"
	default:
		a.l.Panic("unhandled type of agent", a.params.Type)
		return ""
	}
}
