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
	"time"

	"github.com/google/uuid"
)

const defaultTimeout = time.Duration(time.Second * 10)

// ActionResult represents an action result.
type ActionResult struct {
	ID             uuid.UUID
	Name           string
	Error          error
	CombinedOutput []byte
}

// Runner represents action runner.
// Action runner is component that can run an Actions.
type Runner struct {
	out chan ActionResult
}

// NewRunner returns new runner.
func NewRunner() *Runner {
	return &Runner{}
}

// Run runs an Action in separate goroutine.
// When action is ready those output writes to ActionResult channel.
// You can get all action results with ActionReady() method.
func (r *Runner) Run(a Action) {
	go run(a, r.out)
}

// ActionReady returns channel that you can use to read action results.
func (r *Runner) ActionReady() chan ActionResult {
	return r.out
}

func run(a Action, out chan ActionResult) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return
	default:
		cOut, err := a.Run(ctx)
		out <- ActionResult{
			ID:             a.ID(),
			Name:           a.Name(),
			Error:          err,
			CombinedOutput: cOut,
		}
	}
}
