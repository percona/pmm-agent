// pmm-agent
// Copyright 2019 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jobs

import (
	"context"
	"runtime/pprof"
	"sync"

	"github.com/sirupsen/logrus"
)

type Runner struct {
	l *logrus.Entry

	sender Sender

	jobs        chan Job
	runningJobs sync.WaitGroup

	rw         sync.RWMutex
	jobsCancel map[string]context.CancelFunc
}

func NewRunner(sender Sender) *Runner {
	return &Runner{
		l:          logrus.WithField("component", "jobs-executor"),
		sender:     sender,
		jobs:       make(chan Job, 32), // TODO add constant
		jobsCancel: make(map[string]context.CancelFunc),
	}
}

func (e *Runner) Run(ctx context.Context) {
	for {
		select {
		case job := <-e.jobs:
			jobID, jobType := job.ID(), job.Type()

			var ctx context.Context
			var cancel context.CancelFunc
			if timeout := job.Timeout(); timeout != 0 {
				ctx, cancel = context.WithTimeout(ctx, timeout)
			} else {
				ctx, cancel = context.WithCancel(ctx)
			}

			e.addJobCancel(jobID, cancel)
			e.runningJobs.Add(1)
			run := func(ctx context.Context) {
				defer e.runningJobs.Done()
				defer cancel()

				l := e.l.WithFields(logrus.Fields{"id": jobID, "type": jobType})
				l.Infof("Starting...")

				job.Run(ctx, e.sender)
			}

			go pprof.Do(ctx, pprof.Labels("jobID", jobID, "type", jobType), run)
		case <-ctx.Done():
			e.runningJobs.Wait() // wait for all jobs termination
			return
		}
	}
}

func (e *Runner) Start(job Job) {
	e.jobs <- job
}

// Stop stops running Job.
func (e *Runner) Stop(id string) {
	e.rw.RLock()
	defer e.rw.RUnlock()
	if cancel, ok := e.jobsCancel[id]; ok {
		cancel()
	}
}

// IsRunning returns true if job with given ID still running.
func (e *Runner) IsRunning(id string) bool {
	e.rw.RLock()
	defer e.rw.RUnlock()
	_, ok := e.jobsCancel[id]

	return ok
}

func (e *Runner) addJobCancel(jobID string, cancel context.CancelFunc) {
	e.rw.Lock()
	defer e.rw.Unlock()
	e.jobsCancel[jobID] = cancel
}

func (e *Runner) removeJobCancel(jobID string) {
	e.rw.Lock()
	defer e.rw.Unlock()
	delete(e.jobsCancel, jobID)
}
