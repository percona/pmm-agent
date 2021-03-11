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

package channel

import (
	"sync"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"github.com/percona/pmm/api/jobspb"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// JobsChannel encapsulates two-way communication channel between pmm-managed and pmm-agent for jobs execution.
//
// All exported methods are thread-safe.
type JobsChannel struct { //nolint:maligned
	l *logrus.Entry

	streamM sync.Mutex
	stream  jobspb.Jobs_ConnectClient

	mRecv, mSend prometheus.Counter

	requests chan *jobspb.ServerMessage

	closeOnce sync.Once
	closeWait chan struct{}
	closeErr  error
}

// NewJobsChannel creates new two-way communication channel with given stream.
//
// Stream should not be used by the caller after channel is created.
func NewJobsChannel(stream jobspb.Jobs_ConnectClient) *JobsChannel {
	s := &JobsChannel{
		stream: stream,
		l:      logrus.WithField("component", "job_channel"), // only for debug logging

		mRecv: prometheus.NewCounter(prometheus.CounterOpts{ // TODO rename?
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "job_messages_received_total",
			Help:      "A total number of received job messages from pmm-managed.",
		}),
		mSend: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "job_messages_sent_total",
			Help:      "A total number of sent job messages to pmm-managed.",
		}),

		requests: make(chan *jobspb.ServerMessage),

		closeWait: make(chan struct{}),
	}

	go s.runJobsReceiver()
	return s
}

// Wait blocks until channel is closed and returns the reason why it was closed.
//
// When Wait returns, underlying gRPC connection should be terminated to prevent goroutine leak.
func (c *JobsChannel) Wait() error {
	<-c.closeWait
	return c.closeErr
}

// Requests returns a channel for incoming requests. It must be read. It is closed on any error (see Wait).
func (c *JobsChannel) Requests() <-chan *jobspb.ServerMessage {
	return c.requests
}

// Send sends message to pmm-managed.
func (c *JobsChannel) Send(msg *jobspb.AgentMessage) {
	c.streamM.Lock()
	select {
	case <-c.closeWait:
		c.streamM.Unlock()
		return
	default:
	}

	// do not use default compact representation for large/complex messages
	if size := proto.Size(msg); size < 100 {
		c.l.Debugf("Sending message (%d bytes): %s.", size, msg)
	} else {
		c.l.Debugf("Sending message (%d bytes):\n%s\n", size, proto.MarshalTextString(msg))
	}

	err := c.stream.Send(msg)
	c.streamM.Unlock()
	if err != nil {
		c.close(errors.Wrap(err, "failed to send jobs message"))
		return
	}
	c.mSend.Inc()
}

// runReader receives messages from server
func (c *JobsChannel) runJobsReceiver() {
	defer func() {
		close(c.requests)
		c.l.Debug("Exiting receiver goroutine.")
	}()

	for {
		msg, err := c.stream.Recv()
		if err != nil {
			c.close(errors.Wrap(err, "failed to receive message"))
			return
		}
		c.mRecv.Inc()

		// do not use default compact representation for large/complex messages
		if size := proto.Size(msg); size < 100 {
			c.l.Debugf("Received message (%d bytes): %s.", size, msg)
		} else {
			c.l.Debugf("Received message (%d bytes):\n%s\n", size, proto.MarshalTextString(msg))
		}

		c.requests <- msg
	}
}

// close marks channel as closed with given error - only once.
func (c *JobsChannel) close(err error) {
	c.closeOnce.Do(func() {
		c.l.Debugf("Closing with error: %+v", err)
		c.closeErr = err

		c.streamM.Lock()
		_ = c.stream.CloseSend()
		close(c.closeWait)
		c.streamM.Unlock()
	})
}

// Describe implements prometheus.Collector.
func (c *JobsChannel) Describe(ch chan<- *prometheus.Desc) {
	c.mRecv.Describe(ch)
	c.mSend.Describe(ch)
}

// Collect implement prometheus.Collector.
func (c *JobsChannel) Collect(ch chan<- prometheus.Metric) {
	c.mRecv.Collect(ch)
	c.mSend.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*ActionsChannel)(nil)
)
