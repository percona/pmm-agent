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
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/jobspb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type echoJob struct {
	id      string
	timeout time.Duration
	l       *logrus.Entry

	message string
	delay   time.Duration
}

func NewEchoJob(id string, timeout time.Duration, message string, delay time.Duration) Job {
	return &echoJob{
		id:      id,
		timeout: timeout,
		l:       logrus.WithFields(logrus.Fields{"id": id, "type": "echo"}),
		message: message,
		delay:   delay,
	}
}

func (j *echoJob) ID() string {
	return j.id
}

func (j *echoJob) Type() string {
	return "echo"
}

func (j *echoJob) Timeout() time.Duration {
	return j.timeout
}

func (j *echoJob) Run(ctx context.Context, sender Sender) {
	sender.Send(&jobspb.AgentMessage{
		JobId:  j.id,
		Status: status.New(codes.OK, "").Proto(),
		Payload: &jobspb.AgentMessage_JobProgress{
			JobProgress: &jobspb.JobProgress{
				Timestamp: ptypes.TimestampNow(),
				Result: &jobspb.JobProgress_Echo_{
					Echo: &jobspb.JobProgress_Echo{
						Status: fmt.Sprintf("Echo job %s started", j.id),
					},
				},
			},
		},
	})
	delay := time.NewTimer(j.delay)
	defer delay.Stop()

	select {
	case <-delay.C:
		sender.Send(&jobspb.AgentMessage{
			JobId:  j.id,
			Status: status.New(codes.OK, "").Proto(),
			Payload: &jobspb.AgentMessage_JobResult{
				JobResult: &jobspb.JobResult{
					Timestamp: ptypes.TimestampNow(),
					Result: &jobspb.JobResult_Echo_{
						Echo: &jobspb.JobResult_Echo{
							Message: j.message,
						},
					},
				},
			},
		})
	case <-ctx.Done():
		sender.Send(&jobspb.AgentMessage{
			JobId:  j.id,
			Status: status.New(codes.OK, "").Proto(),
			Payload: &jobspb.AgentMessage_JobResult{
				JobResult: &jobspb.JobResult{
					Timestamp: ptypes.TimestampNow(),
					Result: &jobspb.JobResult_Error_{
						Error: &jobspb.JobResult_Error{
							Message: ctx.Err().Error(),
						},
					},
				},
			},
		})
		j.l.Warnf("Job %s terminated: +%v", j.id, ctx.Err())
	}
}
