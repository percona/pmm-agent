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
	id         string
	name       int32
	command    string
	customPath string
	arg        []string

	forbidden map[string]struct{}

	mx     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewShellAction creates Shell action.
//
// Shell action, it's an abstract action that can run some "predefined" set of shell commands.
// This commands can be a shell script, script written on interpreted language, or binary file.
func NewShellAction(id string, command string, params []string) Action {
	return &shellAction{
		id:      id,
		command: command,
		arg:     params,
		forbidden: map[string]struct{}{
			"rm":      {},
			"bash":    {},
			"sh":      {},
			"sudo":    {},
			"unknown": {},
		},
	}
}

// ID returns unique action id.
func (p *shellAction) ID() string {
	return p.id
}

func (p *shellAction) Name() string {
	return p.command
}

// Run runs an action.
func (p *shellAction) Run(ctx context.Context) ([]byte, error) {
	p.mx.Lock()
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.mx.Unlock()

	if _, ok := p.forbidden[p.command]; ok {
		return nil, errUnknownAction
	}

	cmd := exec.CommandContext(ctx, p.command, p.arg...) //nolint:gosec
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return stdoutStderr, nil
}

func (p *shellAction) Stop() {
	p.mx.Lock()
	p.cancel()
	p.mx.Unlock()
}
