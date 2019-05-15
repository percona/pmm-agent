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
//nolint:unused
type ActionResult struct {
	ID             string
	Name           string
	Error          error
	CombinedOutput []byte
}

// ConcurrentRunner represents concurrent action runner.
// Action runner is component that can run an Actions.
//nolint:unused
type ConcurrentRunner struct {
	out    chan ActionResult
	logger logrus.FieldLogger

	// runningActions stores CancelFunc's for running actions.
	runningActions sync.Map // map[string]CancelFunc
	timeout        time.Duration
}

// NewConcurrentRunner returns new runner.
// If timeout is 0 it sets to defaultTimeout constant (10sec).
func NewConcurrentRunner(l logrus.FieldLogger, timeout time.Duration) *ConcurrentRunner {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &ConcurrentRunner{
		logger:  l,
		timeout: timeout,
		out:     make(chan ActionResult),
	}
}

// Run runs an Action in separate goroutine.
// When action is ready those output writes to ActionResult channel.
// You can get all action results with ActionReady() method.
func (r *ConcurrentRunner) Run(a Action) {
	go r.run(a, r.timeout)
}

// ActionReady returns channel that you can use to read action results.
func (r *ConcurrentRunner) ActionReady() <-chan ActionResult {
	return r.out
}

// Stop stops running action.
func (r *ConcurrentRunner) Stop(id string) {
	if a, ok := r.runningActions.Load(id); ok {
		if cancel, ok := a.(context.CancelFunc); ok {
			cancel()
			r.runningActions.Delete(id)
		}
	}
}

func (r *ConcurrentRunner) run(a Action, t time.Duration) { //nolint:unused
	tCtx, tCancel := context.WithTimeout(context.Background(), t)
	ctx, cancel := context.WithCancel(tCtx)
	defer tCancel()

	r.runningActions.Store(a.ID(), cancel)
	l := r.logger.WithFields(logrus.Fields{"id": a.ID(), "name": a.Name()})
	l.Debugf("Running action...")

	cOut, err := a.Run(ctx)
	r.runningActions.Delete(a.ID())
	l.Debugf("Action finished")

	r.out <- ActionResult{
		ID:             a.ID(),
		Error:          err,
		CombinedOutput: cOut,
	}
}
