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
	"path"

	"github.com/google/uuid"
)

type ptSummary struct {
	id         uuid.UUID
	name       string
	customPath string
	params     map[string]string
}

// NewPtSummary creates pt-summary action
func NewPtSummary(params map[string]string, customPath string) Action {
	return &ptSummary{
		id:         uuid.New(),
		name:       "pt-summary",
		params:     params,
		customPath: customPath,
	}
}

// ID returns unique action id.
func (p *ptSummary) ID() uuid.UUID {
	return p.id
}

// ID returns action name.
func (p *ptSummary) Name() string {
	return p.name
}

// Run runs an action.
func (p *ptSummary) Run(ctx context.Context) ([]byte, error) {
	executable := path.Join(p.customPath, p.name)
	cmd := exec.CommandContext(ctx, executable, parseArguments(p.params)...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return stdoutStderr, nil
}
