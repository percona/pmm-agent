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

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentpb"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/client/channel"
)

// Sender provides method for sending message to PMM server.
type Sender interface {
	SendResponse(msg *channel.AgentResponse)
}

// Runner allows to execute jobs.
type Runner struct {
	l *logrus.Entry

	sender Sender

	jobs        chan Job
	runningJobs sync.WaitGroup

	rw         sync.RWMutex
	jobsCancel map[string]context.CancelFunc
}

// NewRunner creates new jobs runner.
func NewRunner(sender Sender) *Runner {
	return &Runner{
		l:          logrus.WithField("component", "jobs-executor"),
		sender:     sender,
		jobs:       make(chan Job, 32), // TODO add constant
		jobsCancel: make(map[string]context.CancelFunc),
	}
}

// Run starts jobs execution loop. It reads jobs from the channel and starts them in separate goroutines.
func (r *Runner) Run(ctx context.Context) {
	for {
		select {
		case job := <-r.jobs:
			jobID, jobType := job.ID(), job.Type()

			var nCtx context.Context
			var cancel context.CancelFunc
			if timeout := job.Timeout(); timeout != 0 {
				nCtx, cancel = context.WithTimeout(ctx, timeout)
			} else {
				nCtx, cancel = context.WithCancel(ctx)
			}

			r.addJobCancel(jobID, cancel)
			r.runningJobs.Add(1)
			run := func(ctx context.Context) {
				defer r.runningJobs.Done()
				defer cancel()

				l := r.l.WithFields(logrus.Fields{"id": jobID, "type": jobType})
				l.Infof("Starting...")

				err := job.Run(ctx, r.send)
				if err != nil {
					r.sender.SendResponse(&channel.AgentResponse{
						Payload: &agentpb.JobResult{
							JobId:     job.ID(),
							Timestamp: ptypes.TimestampNow(),
							Result: &agentpb.JobResult_Error_{
								Error: &agentpb.JobResult_Error{
									Message: err.Error(),
								},
							},
						},
					})
					l.Warnf("Job terminated with error %+v", err)
				}
			}

			go pprof.Do(nCtx, pprof.Labels("jobID", jobID, "type", jobType), run)
		case <-ctx.Done():
			r.runningJobs.Wait() // wait for all jobs termination
			return
		}
	}
}

func (r *Runner) send(payload agentpb.AgentResponsePayload) {
	r.sender.SendResponse(&channel.AgentResponse{
		ID:      0, // TODO add comment
		Payload: payload,
	})
}

// Start starts given job.
func (r *Runner) Start(job Job) {
	r.jobs <- job
}

// Stop stops running Job.
func (r *Runner) Stop(id string) {
	r.rw.RLock()
	defer r.rw.RUnlock()
	if cancel, ok := r.jobsCancel[id]; ok {
		cancel()
	}
}

// IsRunning returns true if job with given ID still running.
func (r *Runner) IsRunning(id string) bool {
	r.rw.RLock()
	defer r.rw.RUnlock()
	_, ok := r.jobsCancel[id]

	return ok
}

func (r *Runner) addJobCancel(jobID string, cancel context.CancelFunc) {
	r.rw.Lock()
	defer r.rw.Unlock()
	r.jobsCancel[jobID] = cancel
}

func (r *Runner) removeJobCancel(jobID string) {
	r.rw.Lock()
	defer r.rw.Unlock()
	delete(r.jobsCancel, jobID)
}
