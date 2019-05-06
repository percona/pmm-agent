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
	"errors"

	"github.com/google/uuid"
)

// Action describe abstract action that can be runned and must return []bytes slice.
// Every structure that implement this interface can be an action.
type Action interface {
	ID() uuid.UUID
	Name() string
	// Run runs command and returns output and error.
	// This method is blocking.
	Run(ctx context.Context) ([]byte, error)
}

func New(name string, params map[string]string) (Action, error) {
	switch name {
	case "pt-summary":
		return NewPtSummary(params, ""), nil
	}
	return nil, errors.New("unsupported action")
}

func parseArguments(params map[string]string) []string {
	args := make([]string, 0)
	for k, v := range params {
		args = append(args, k)
		if len(v) > 0 {
			args = append(args, v)
		}
	}
	return args
}
