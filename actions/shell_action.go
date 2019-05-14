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

package actions

import (
	"context"
	"os/exec"
	"sync"
)

type shellAction struct {
	id      string
	command string
	arg     []string

	// Because cxt, and cancel are using in Run() which is blocked and Stop(),
	// and because those methods can be used by separate goroutine
	// we should protect those vars by Mutex.
	mx     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewShellAction creates Shell action.
//
// Shell action, it's an abstract action that can run an external commands.
// This commands can be a shell script, script written on interpreted language, or binary file.
func NewShellAction(id string, cmd string, arg []string) Action {
	return &shellAction{
		id:      id,
		command: cmd,
		arg:     arg,
	}
}

// ID returns unique action id.
func (p *shellAction) ID() string {
	return p.id
}

// Name returns action name as as string.
func (p *shellAction) Name() string {
	return p.command
}

// Run starts an action.
// Action runs with internal CancelContext (by default), and can be stopped with Stop() method.
// This method is blocking.
func (p *shellAction) Run(ctx context.Context) ([]byte, error) {
	p.mx.Lock()
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.mx.Unlock()

	cmd := exec.CommandContext(p.ctx, p.command, p.arg...) //nolint:gosec
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return stdoutStderr, nil
}

// Stop stops started action.
// Calls cancel() of internal cation context.
// If action isn't started yet, returns "false".
func (p *shellAction) Stop() bool {
	p.mx.Lock()
	defer p.mx.Unlock()
	if p.cancel != nil {
		p.cancel()
		return true
	}
	return false
}
