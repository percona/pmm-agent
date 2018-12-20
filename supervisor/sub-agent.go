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
	"math"
	"math/rand"
	"os"
	"os/exec"
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
	log         *logger.CircularWriter
	l           *logrus.Entry
	params      *agent.SetStateRequest_AgentProcess
	port        uint32
	cmd         *exec.Cmd
	state       *fsm.FSM
	changesChan chan string
}

func New(ctx context.Context, params *agent.SetStateRequest_AgentProcess) *subAgent {
	l := logrus.WithField("component", "runner").
		WithField("agentID", params.AgentId).
		WithField("type", params.Type)

	changesChan := make(chan string, 10) // TODO: writing into changesChan blocks logic.

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
		params:      params,
		log:         logger.New(10),
		l:           l,
		state:       state,
		changesChan: changesChan,
	}
	go sAgent.start(ctx)
	return sAgent
}

// start starts sub-agent.
func (m *subAgent) start(ctx context.Context) {
	restartCount := 0

	for m.state.Can(START) {
		err := m.state.Event(START)
		if err != nil {
			return
		}
		filename := m.binary()
		if m.exec(ctx, filename) {
			restartCount = 0
		}
		max := math.Pow(2, float64(restartCount))
		delay := rand.Int63n(int64(max))
		startTime := time.After(time.Duration(delay) * time.Millisecond)
		err = m.state.Event(RESTART)
		if err != nil {
			m.l.Debug(err)
			return
		}
	L:
		for {
			select {
			case <-ctx.Done():
				err = m.state.Event(EXIT)
				if err != nil {
					m.l.Debug(err)
				}
				return
			case <-startTime:
				restartCount++
				break L
			}
		}
	}
}

func (m *subAgent) exec(ctx context.Context, filename string) bool {
	cmd := exec.CommandContext(ctx, filename, m.params.Args...)
	cmd.Env = m.params.Env
	cmd.Stdout = io.MultiWriter(os.Stdout, m.log)
	cmd.Stderr = io.MultiWriter(os.Stderr, m.log)
	err := cmd.Start()
	if err != nil {
		return false
	}
	m.cmd = cmd
	time.Sleep(2 * time.Second)
	if !m.processExists() {
		return false
	}
	err = m.state.Event(STARTED)
	if err != nil {
		return false
	}
	err = m.cmd.Wait()
	if err != nil {
		m.l.Debug(err)
	}
	return true
}

// stop stops sub-agent
func (m *subAgent) Stop() {
	go func() {
		err := m.state.Event(STOP)
		if err != nil {
			m.l.Warnln("Can't change state to STOPPING")
			return
		}

		err = m.cmd.Process.Signal(syscall.SIGINT)
		if err != nil {
			m.l.Warnln("Can't stop sub-agent")
			return
		}

		time.Sleep(5 * time.Second)

		if m.processExists() {
			err := m.cmd.Process.Kill()
			if err != nil {
				m.l.Warnln("Can't kill sub-agent")
				return
			}
		}

		err = m.state.Event(EXIT)
		if err != nil {
			m.l.Warnln("Can't change state to STOPPING")
			return
		}
	}()
}

func (m *subAgent) processExists() bool {
	return syscall.Kill(m.cmd.Process.Pid, syscall.Signal(0)) == nil
}

// GetLogs returns logs from sub-agent STDOut and STDErr.
func (m *subAgent) GetLogs() []string {
	return m.log.Data()
}

func (m *subAgent) pid() int {
	if m.state.Is(RUNNING) {
		return m.cmd.Process.Pid
	}
	return 0
}

func (m *subAgent) Changes() <-chan string {
	return m.changesChan
}
