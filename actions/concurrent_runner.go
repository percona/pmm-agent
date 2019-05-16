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
	ID             string
	Type           string
	Error          error
	CombinedOutput []byte
}

// ConcurrentRunner represents concurrent action runner.
// Action runner is component that can run an Actions.
//nolint:unused
type ConcurrentRunner struct {
	runningActions sync.WaitGroup
	out            chan ActionResult
	l              *logrus.Entry

	mx            sync.Mutex
	actionsCancel map[string]context.CancelFunc

	timeout time.Duration
	appCtx  context.Context
}

// NewConcurrentRunner returns new runner.
// If timeout is 0 it sets to defaultTimeout constant (10sec).
func NewConcurrentRunner(appCtx context.Context, l *logrus.Entry, timeout time.Duration) *ConcurrentRunner {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	r := &ConcurrentRunner{
		appCtx:        appCtx,
		l:             l,
		timeout:       timeout,
		out:           make(chan ActionResult),
		actionsCancel: make(map[string]context.CancelFunc),
	}

	go func() {
		<-appCtx.Done()
		r.runningActions.Wait()
		close(r.out)
	}()

	return r
}

// Start runs an Action in separate goroutine.
// When action is ready those output writes to ActionResult channel.
// You can get all action results with ActionReady() method.
func (r *ConcurrentRunner) Start(a action) {
	r.runningActions.Add(1)
	go func() {
		defer r.runningActions.Done()
		tCtx, tCancel := context.WithTimeout(r.appCtx, r.timeout)
		ctx, cancel := context.WithCancel(tCtx)
		defer tCancel()

		r.mx.Lock()
		r.actionsCancel[a.ID()] = cancel
		r.mx.Unlock()

		l := r.l.WithFields(logrus.Fields{"id": a.ID(), "type": a.Type()})
		l.Debugf("Running action...")

		cOut, err := a.Run(ctx)

		r.mx.Lock()
		delete(r.actionsCancel, a.ID())
		r.mx.Unlock()

		l.Debugf("Action finished")

		r.out <- ActionResult{
			ID:             a.ID(),
			Error:          err,
			CombinedOutput: cOut,
		}
	}()
}

// ActionReady returns channel that you can use to read action results.
func (r *ConcurrentRunner) ActionReady() <-chan ActionResult {
	return r.out
}

// Stop stops running action.
func (r *ConcurrentRunner) Stop(id string) {
	r.mx.Lock()
	defer r.mx.Unlock()
	if cancel, ok := r.actionsCancel[id]; ok {
		cancel()
	}
}
