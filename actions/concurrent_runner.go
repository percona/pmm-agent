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
	"runtime/pprof"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const defaultTimeout = time.Second * 10

var (
	errChannelClosed = errors.New("Actions channel was closed")
)

// ActionResult represents an Action result.
type ActionResult struct {
	ID     string
	Output []byte
	Error  error
}

// ConcurrentRunner represents concurrent Action runner.
// Action runner is component that can run an Actions.
type ConcurrentRunner struct {
	ctx     context.Context
	l       *logrus.Entry
	timeout time.Duration
	results chan ActionResult

	runningActions sync.WaitGroup

	rw            sync.RWMutex
	actionsCancel map[string]context.CancelFunc
}

// NewConcurrentRunner returns new runner.
// With this component you can run actions concurrently and read action results when they will be finished.
// If timeout is 0 it sets to default = 10 seconds.
func NewConcurrentRunner(ctx context.Context, l *logrus.Entry, timeout time.Duration) *ConcurrentRunner {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	r := &ConcurrentRunner{
		ctx:           ctx,
		l:             l,
		timeout:       timeout,
		results:       make(chan ActionResult),
		actionsCancel: make(map[string]context.CancelFunc),
	}

	// When an external context is done, we waiting for all running actions to finish and then closing "r.out" channel.
	// The reason we doing this is to guarantee, all actions will return its output data
	// and only then method "NextActionResult()" will return an error.
	go func() {
		<-ctx.Done()
		r.runningActions.Wait()
		close(r.results)
	}()

	return r
}

// Start runs an Action in separate goroutine.
// Call of this method doesn't block execution.
// When Action will be ready you can read it result by WaitNextAction() method.
func (r *ConcurrentRunner) Start(a Action) {
	if err := r.ctx.Err(); err != nil {
		r.l.Errorf("Ignoring Start: %s.", err)
		return
	}

	actionID, actionType := a.ID(), a.Type()
	r.runningActions.Add(1)
	ctx, cancel := context.WithTimeout(r.ctx, r.timeout)
	run := func(ctx context.Context) {
		defer r.runningActions.Done()
		defer cancel()

		r.rw.Lock()
		r.actionsCancel[actionID] = cancel
		r.rw.Unlock()

		l := r.l.WithFields(logrus.Fields{"id": actionID, "type": actionType})
		l.Debugf("Running Action...")

		b, err := a.Run(ctx)

		r.rw.Lock()
		delete(r.actionsCancel, actionID)
		r.rw.Unlock()

		l.Debugf("Action finished")

		r.results <- ActionResult{
			ID:     actionID,
			Output: b,
			Error:  err,
		}
	}
	go pprof.Do(ctx, pprof.Labels("actionID", actionID, "type", actionType), run)
}

// WaitNextAction returns an action result.
// Calling this method blocks execution and wait for next action will be finished.
// Each time the action becomes finished method returns an action result.
// The error will be returned after all actions were finished and when the runner is going to stop their work.
func (r *ConcurrentRunner) WaitNextAction() (ActionResult, error) {
	ar, ok := <-r.results
	if !ok {
		return ar, errChannelClosed
	}
	return ar, nil
}

// Stop stops running Action.
func (r *ConcurrentRunner) Stop(id string) {
	r.rw.RLock()
	defer r.rw.RUnlock()
	if cancel, ok := r.actionsCancel[id]; ok {
		cancel()
	}
}
