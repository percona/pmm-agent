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

// Package client contains business logic of working with pmm-managed.
package client

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/version"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/percona/pmm-agent/client/channel"
	"github.com/percona/pmm-agent/config"
	"github.com/percona/pmm-agent/utils/backoff"
)

const (
	dialTimeout       = 5 * time.Second
	backoffMinDelay   = 1 * time.Second
	backoffMaxDelay   = 15 * time.Second
	clockDriftWarning = 5 * time.Second
)

// supervisor is a subset of methods of supervisor.Supervisor used by this package.
// We use it instead of real type for testing and to avoid dependency cycle.
type supervisor interface {
	Changes() <-chan agentpb.StateChangedRequest
	QANRequests() <-chan agentpb.QANCollectRequest
	SetState(*agentpb.SetStateRequest)
}

// Client represents pmm-agent's connection to nginx/pmm-managed.
type Client struct {
	cfg        *config.Config
	supervisor supervisor

	l       *logrus.Entry
	backoff *backoff.Backoff

	rw      sync.RWMutex
	channel *channel.Channel
}

// New creates new client.
//
// Caller should call Run.
func New(cfg *config.Config, supervisor supervisor) *Client {
	return &Client{
		cfg:        cfg,
		supervisor: supervisor,
		l:          logrus.WithField("component", "client"),
		backoff:    backoff.New(backoffMinDelay, backoffMaxDelay),
	}
}

// Run connects to the server, processes requests and sends responses.
//
// Once Run exits, connection is closed, and caller should stop supervisor.
// That Client instance can't be reused ofter that.
func (c *Client) Run(ctx context.Context) error {
	c.l.Info("Starting...")
	defer c.l.Info("Done.")

	// do nothing until ctx is canceled if address is not given
	if c.cfg.Address == "" {
		c.l.Error("PMM Server address is not provided, halting.")
		<-ctx.Done()
		return errors.Wrap(ctx.Err(), "no address")
	}

	// try to connect until success, or until ctx is canceled
	var dialResult *dialResult
	for {
		dialCtx, dialCancel := context.WithTimeout(ctx, dialTimeout)
		dialResult = dial(dialCtx, c.cfg, c.l)
		dialCancel()
		if dialResult != nil {
			break
		}

		retryCtx, retryCancel := context.WithTimeout(ctx, c.backoff.Delay())
		<-retryCtx.Done()
		retryCancel()
		if ctx.Err() != nil {
			break
		}
	}
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "failed to connect")
	}

	defer func() {
		switch err := dialResult.conn.Close(); err {
		case nil:
			c.l.Info("Connection closed.")
		default:
			c.l.Errorf("Connection closed: %s.", err)
		}
	}()

	c.rw.Lock()
	c.channel = dialResult.channel
	c.rw.Unlock()

	// TODO set metadata

	// Once the client is connected, ctx cancellation is ignored.
	// We start two goroutines, and terminate the gRPC connection and leave Run when any of them finished:
	// 1. processSupervisorRequests reads requests (status changes and QAN data) from the supervisor and sends them to the channel.
	//    It finishes when the supervisor is stopped.
	//    When the gRPC connection is terminated on leaving Run, processChannelRequests exits too.
	// 2. processChannelRequests reads requests from the channel and processes them.
	//    It finishes when an unexpected message is received from the channel, or when can't be received at all.
	//    When Run is left, caller stops supervisor, and that allows processSupervisorRequests to exit.
	done := make(chan error, 2)
	go func() {
		done <- c.processSupervisorRequests()
	}()
	go func() {
		done <- c.processChannelRequests()
	}()
	err := <-done
	c.l.Error(err)
	return err
}

func (c *Client) processSupervisorRequests() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for state := range c.supervisor.Changes() {
			res := c.channel.SendRequest(&agentpb.AgentMessage_StateChanged{
				StateChanged: &state,
			})
			if res == nil {
				c.l.Warn("Failed to send StateChanged request.")
			}
		}
		c.l.Info("Supervisor changes done.")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for collect := range c.supervisor.QANRequests() {
			res := c.channel.SendRequest(&agentpb.AgentMessage_QanCollect{
				QanCollect: &collect,
			})
			if res == nil {
				c.l.Warn("Failed to send QanCollect request.")
			}
		}
		c.l.Info("Supervisor QAN requests done.")
	}()

	wg.Wait()
	return errors.New("supervisor done")
}

func (c *Client) processChannelRequests() error {
	for serverMessage := range c.channel.Requests() {
		var agentMessage *agentpb.AgentMessage
		switch payload := serverMessage.Payload.(type) {
		case *agentpb.ServerMessage_Ping:
			agentMessage = &agentpb.AgentMessage{
				Id: serverMessage.Id,
				Payload: &agentpb.AgentMessage_Pong{
					Pong: &agentpb.Pong{
						CurrentTime: ptypes.TimestampNow(),
					},
				},
			}

		case *agentpb.ServerMessage_SetState:
			c.supervisor.SetState(payload.SetState)

			agentMessage = &agentpb.AgentMessage{
				Id: serverMessage.Id,
				Payload: &agentpb.AgentMessage_SetState{
					SetState: new(agentpb.SetStateResponse),
				},
			}

		default:
			// Requests() is not closed, so exit early to break channel
			c.l.Errorf("Unhandled server message payload: %s.", payload)
			return errors.New("unhandled payload")
		}

		c.channel.SendResponse(agentMessage)
	}

	// once Requests() is closed, caller can learn why
	return c.channel.Wait()
}

type dialResult struct {
	conn         *grpc.ClientConn
	streamCancel context.CancelFunc
	md           agentpb.AgentServerMetadata
	channel      *channel.Channel
}

// dial tries to connect to the server once.
func dial(dialCtx context.Context, cfg *config.Config, l *logrus.Entry) *dialResult {
	host, _, _ := net.SplitHostPort(cfg.Address)
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: cfg.InsecureTLS, //nolint:gosec
	}
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithUserAgent("pmm-agent/" + version.Version),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}

	l.Infof("Connecting to %s ...", cfg.Address)
	conn, err := grpc.DialContext(dialCtx, cfg.Address, opts...)
	if err != nil {
		msg := err.Error()

		// improve error message in that particular case
		if err == context.DeadlineExceeded {
			msg = "connection timeout"
		}

		l.Errorf("Failed to connect to %s: %s.", cfg.Address, msg)
		return nil
	}
	l.Infof("Connected to %s.", cfg.Address)

	streamCtx, streamCancel := context.WithCancel(context.Background())
	teardown := func() {
		streamCancel()
		err = conn.Close()
		if err == nil {
			l.Debugf("Connection closed.")
		} else {
			l.Debugf("Connection closed with %s.", err)
		}
	}

	l.Info("Establishing two-way communication channel ...")
	streamCtx = agentpb.AddAgentConnectMetadata(streamCtx, &agentpb.AgentConnectMetadata{
		ID:      cfg.ID,
		Version: version.Version,
	})
	stream, err := agentpb.NewAgentClient(conn).Connect(streamCtx)
	if err != nil {
		l.Errorf("Failed to establish two-way communication channel: %s.", err)
		teardown()
		return nil
	}

	md, err := agentpb.GetAgentServerMetadata(stream)
	if err != nil {
		// TODO make it a hard error
		l.Warnf("Can't get server metadata: %s.", err)
	}

	// So far nginx can handle all that itself without pmm-managed.
	// We need to send ping to ensure that pmm-managed is alive and that Agent ID is valid.
	start := time.Now()
	channel := channel.NewChannel(stream)
	res := channel.SendRequest(&agentpb.AgentMessage_Ping{
		Ping: new(agentpb.Ping),
	})
	if res == nil {
		err = channel.Wait()
		msg := err.Error()

		// improve error message in that particular case
		status := status.Convert(errors.Cause(err))
		if status.Code() == codes.Internal && strings.Contains(status.Message(), "received the unexpected content-type") {
			msg += "\nPlease check that pmm-managed is running"
		}

		l.Errorf("Failed to send Ping message: %s.", msg)
		teardown()
		return nil
	}

	roundtrip := time.Since(start)
	serverTime, err := ptypes.Timestamp(res.(*agentpb.ServerMessage_Pong).Pong.CurrentTime)
	if err != nil {
		l.Errorf("Failed to decode Pong.current_time: %s.", err)
		teardown()
		return nil
	}
	l.Infof("Two-way communication channel established in %s.", roundtrip)

	clockDrift := serverTime.Sub(start) - roundtrip/2
	if clockDrift > clockDriftWarning || -clockDrift > clockDriftWarning {
		l.Warnf("Estimated clock drift: %s.", clockDrift)
	}

	return &dialResult{conn, streamCancel, md, channel}
}

// Describe implements prometheus.Collector.
func (c *Client) Describe(chan<- *prometheus.Desc) {
	// Sending no descriptor at all marks the Collector as “unchecked”,
	// i.e. no checks will be performed at registration time, and the
	// Collector may yield any Metric it sees fit in its Collect method.
	return
}

// Collect implements prometheus.Collector.
func (c *Client) Collect(ch chan<- prometheus.Metric) {
	c.rw.RLock()
	channel := c.channel
	c.rw.RUnlock()

	if channel != nil {
		channel.Collect(ch)
	}
}

// check interface
var (
	_ prometheus.Collector = (*Client)(nil)
)
