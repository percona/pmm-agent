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
)

type shellAction struct {
	id      string
	command string
	arg     []string
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

// Run starts an action. This method is blocking.
func (p *shellAction) Run(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, p.command, p.arg...) //nolint:gosec
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return b, nil
}
