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

	"github.com/sirupsen/logrus"
)

const defaultTimeout = time.Second * 10

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
	timeout time.Duration
	l       *logrus.Entry

	rw            sync.RWMutex
	actionsCancel map[string]context.CancelFunc
}

// NewConcurrentRunner returns new runner.
// With this component you can run actions concurrently and read action results when they will be finished.
// If timeout is 0 it sets to default = 10 seconds.
//
// ConcurrentRunner is stopped when context passed to NewConcurrentRunner is canceled.
// Results are reported via Results() channel which must be read until it is closed.
func NewConcurrentRunner(ctx context.Context, timeout time.Duration) *ConcurrentRunner {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	r := &ConcurrentRunner{
		ctx:           ctx,
		timeout:       timeout,
		l:             logrus.WithField("component", "actions-runner"),
		actionsCancel: make(map[string]context.CancelFunc),
	}

	return r
}

// Start starts an Action in a separate goroutine.
// Returns ActionResult channel. You can read ActionResult from it.
// Also returns error if Start was called after Context was Done.
func (r *ConcurrentRunner) Start(a Action) (<-chan *ActionResult, error) {
	res := make(chan *ActionResult)
	if err := r.ctx.Err(); err != nil {
		return nil, err
	}

	actionID, actionType := a.ID(), a.Type()
	ctx, cancel := context.WithTimeout(r.ctx, r.timeout)
	run := func(ctx context.Context) {
		defer close(res) // close should be executed by sender, so here it is.
		defer cancel()

		r.rw.Lock()
		r.actionsCancel[actionID] = cancel
		r.rw.Unlock()

		l := r.l.WithFields(logrus.Fields{"id": actionID, "type": actionType})
		l.Infof("Starting...")

		b, err := a.Run(ctx)

		r.rw.Lock()
		delete(r.actionsCancel, actionID)
		r.rw.Unlock()

		if err == nil {
			l.Infof("Done without error.")
		} else {
			l.Warnf("Done with error: %s.", err)
		}

		res <- &ActionResult{
			ID:     actionID,
			Output: b,
			Error:  err,
		}
	}
	go pprof.Do(ctx, pprof.Labels("actionID", actionID, "type", actionType), run)

	return res, nil
}

// Stop stops running Action.
func (r *ConcurrentRunner) Stop(id string) {
	r.rw.RLock()
	defer r.rw.RUnlock()
	if cancel, ok := r.actionsCancel[id]; ok {
		cancel()
	}
}
