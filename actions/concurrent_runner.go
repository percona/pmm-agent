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

	"github.com/sirupsen/logrus"
)

const defaultTimeout = time.Second * 10

// ActionResult represents an action result.
type ActionResult struct {
	//nolint:unused
	ID             string
	Name           string
	Error          error
	CombinedOutput []byte
}

// ConcurrentRunner represents concurrent action runner.
// Action runner is component that can run an Actions.
type ConcurrentRunner struct {
	//nolint:unused
	out    chan ActionResult
	logger logrus.FieldLogger

	rw      sync.RWMutex
	actions map[string]Action
}

// NewRunner returns new runner.
func NewConcurrentRunner(l logrus.FieldLogger) *ConcurrentRunner {
	return &ConcurrentRunner{
		logger:  l,
		out:     make(chan ActionResult),
		actions: make(map[string]Action),
	}
}

// Run runs an Action in separate goroutine.
// When action is ready those output writes to ActionResult channel.
// You can get all action results with ActionReady() method.
func (r *ConcurrentRunner) Run(a Action) {
	r.rw.Lock()
	r.actions[a.ID()] = a
	r.rw.Unlock()

	go run(a, r.out, r.logger)
}

// ActionReady returns channel that you can use to read action results.
func (r *ConcurrentRunner) ActionReady() <-chan ActionResult {
	return r.out
}

// Stop stops running action.
func (r *ConcurrentRunner) Stop(id string) {
	r.rw.Lock()
	defer r.rw.Unlock()
	if a, ok := r.actions[id]; ok {
		a.Stop()
		delete(r.actions, id)
	}
}

func run(a Action, out chan<- ActionResult, logger logrus.FieldLogger) { //nolint:unused
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	logger.WithFields(logrus.Fields{
		"id":   a.ID(),
		"name": a.Name(),
	}).Debugf("Running action...")

	select {
	case <-ctx.Done():
		logger.WithFields(logrus.Fields{
			"id":   a.ID(),
			"name": a.Name(),
		}).Debugf("Action canceled")

		return
	default:
		cOut, err := a.Run(ctx)

		logger.WithFields(logrus.Fields{
			"id":   a.ID(),
			"name": a.Name(),
		}).Debugf("Action finished")

		out <- ActionResult{
			ID:             a.ID(),
			Error:          err,
			CombinedOutput: cOut,
		}
	}
}
