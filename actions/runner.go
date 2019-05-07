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
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const defaultTimeout = time.Second * 10

// ActionResult represents an action result.
type ActionResult struct { //nolint:unused
	ID             uuid.UUID
	Name           string
	Error          error
	CombinedOutput []byte
}

// Runner represents action runner.
// Action runner is component that can run an Actions.
type Runner struct { //nolint:unused
	out    chan ActionResult
	logger logrus.FieldLogger

	rw      sync.RWMutex
	actions map[uuid.UUID]context.CancelFunc
}

// NewRunner returns new runner.
func NewRunner(l logrus.FieldLogger) *Runner {
	return &Runner{
		logger:  l,
		out:     make(chan ActionResult),
		actions: make(map[uuid.UUID]context.CancelFunc),
	}
}

// Run runs an Action in separate goroutine.
// When action is ready those output writes to ActionResult channel.
// You can get all action results with ActionReady() method.
func (r *Runner) Run(a Action) {
	ctx, cancel := context.WithCancel(context.Background())

	r.rw.Lock()
	r.actions[a.ID()] = cancel
	r.rw.Unlock()

	go run(ctx, a, r.out, r.logger)
}

// ActionReady returns channel that you can use to read action results.
func (r *Runner) ActionReady() <-chan ActionResult {
	return r.out
}

// Cancel stops running action.
func (r *Runner) Cancel(id uuid.UUID) {
	r.rw.Lock()
	defer r.rw.Unlock()
	if cancel, ok := r.actions[id]; ok {
		cancel()
		delete(r.actions, id)
	}
}

func run(ctx context.Context, a Action, out chan<- ActionResult, logger logrus.FieldLogger) { //nolint:unused
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	logger.WithFields(logrus.Fields{
		"action_id":   a.ID(),
		"action_name": a.Name(),
	}).Debugf("Running action...")

	select {
	case <-ctx.Done():
		logger.WithFields(logrus.Fields{
			"action_id":   a.ID(),
			"action_name": a.Name(),
		}).Debugf("Action canceled")

		return
	default:
		cOut, err := a.Run(ctx)

		logger.WithFields(logrus.Fields{
			"action_id":   a.ID(),
			"action_name": a.Name(),
		}).Debugf("Action finished")

		out <- ActionResult{
			ID:             a.ID(),
			Name:           a.Name(),
			Error:          err,
			CombinedOutput: cOut,
		}
	}
}
