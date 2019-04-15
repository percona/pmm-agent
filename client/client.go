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

package client

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/percona/pmm-agent/config"
)

//nolint: unused
const (
	dialTimeout       = 10 * time.Second
	backoffMaxDelay   = 10 * time.Second
	clockDriftWarning = 5 * time.Second
)

//nolint: unused
type stateReceiver interface {
	Changes() <-chan agentpb.StateChangedRequest
	QANRequests() <-chan agentpb.QANCollectRequest
}

//nolint: unused
type stateChanger interface {
	stateReceiver
	SetState(*agentpb.SetStateRequest)
}

//nolint: unused
type metadataReader interface {
	ReadMetadata(md agentpb.AgentServerMetadata)
}

// Client pmm-agent gRPC client implementation.
//nolint: unused
type Client struct {
	logger            logrus.FieldLogger
	appCtx            context.Context
	backoffMaxDelay   time.Duration
	dialTimeout       time.Duration
	clockDriftWarning time.Duration

	conn    *grpc.ClientConn
	cfg     *config.Config
	aClient agentpb.AgentClient
	channel *Channel

	stateCgr stateChanger
	mdReader metadataReader

	streamCancel context.CancelFunc
	done         chan bool
	wg           sync.WaitGroup
}

// New creates new agent.
func New(appCtx context.Context, cfg *config.Config, stateChanger stateChanger, mdReader metadataReader) *Client {
	return &Client{
		appCtx:            appCtx,
		logger:            logrus.WithField("component", "client"),
		backoffMaxDelay:   backoffMaxDelay,
		dialTimeout:       dialTimeout,
		clockDriftWarning: clockDriftWarning,
		stateCgr:          stateChanger,
		mdReader:          mdReader,
		cfg:               cfg,
		done:              make(chan bool),
	}
}

// Wait blocks until client end its work.
func (c *Client) Wait() {
	<-c.done
}

// Run connects agent to gRPC pmm-server
func (c *Client) Run() {
	var err error

	host, _, _ := net.SplitHostPort(c.cfg.Address)
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: c.cfg.InsecureTLS, //nolint:gosec
	}
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithBackoffMaxDelay(c.backoffMaxDelay),
		grpc.WithUserAgent("pmm-agent/" + version.Version),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}

	c.logger.Infof("Connecting to %s ...", c.cfg.Address)
	dialCtx, dialCancel := context.WithTimeout(c.appCtx, c.dialTimeout)
	c.conn, err = grpc.DialContext(dialCtx, c.cfg.Address, opts...)
	dialCancel()
	if err != nil {
		c.logger.Fatalf("Failed to connect to %s: %s.", c.cfg.Address, err)
	}

	c.logger.Infof("Connected to %s.", c.cfg.Address)
	c.aClient = agentpb.NewAgentClient(c.conn)

	// use separate context for stream to cancel it after supervisor is done sending last changes
	var streamCtx context.Context
	streamCtx, c.streamCancel = context.WithCancel(context.Background())
	streamCtx = agentpb.AddAgentConnectMetadata(streamCtx, &agentpb.AgentConnectMetadata{
		ID:      c.cfg.ID,
		Version: version.Version,
	})

	c.logger.Info("Establishing two-way communication channel ...")
	stream, err := c.aClient.Connect(streamCtx)
	if err != nil {
		c.logger.Errorf("Failed to establish two-way communication channel: %s.", err)
		c.streamCancel()
		return
	}

	c.channel = NewChannel(stream)
	prometheus.MustRegister(c.channel)
	c.wg.Add(1)
	go func() {
		err = c.channel.Wait()
		switch err {
		case nil:
			c.logger.Info("Two-way communication channel closed.")
		default:
			c.logger.Errorf("Two-way communication channel closed: %s", err)
		}
		defer c.wg.Done()
	}()

	// So far nginx can handle all that itself without pmm-managed.
	// We need to send ping to ensure that pmm-managed is alive and that Agent ID is valid.
	start := time.Now()
	res := c.channel.SendRequest(&agentpb.AgentMessage_Ping{
		Ping: new(agentpb.Ping),
	})
	if res == nil {
		// error will be logged by channel code
		c.streamCancel()
		return
	}
	roundtrip := time.Since(start)
	serverTime, err := ptypes.Timestamp(res.(*agentpb.ServerMessage_Pong).Pong.CurrentTime)
	if err != nil {
		c.logger.Errorf("Failed to decode Pong.current_time: %s.", err)
		c.streamCancel()
		return
	}
	c.logger.Infof("Two-way communication channel established in %s.", roundtrip)
	clockDrift := serverTime.Sub(start) - roundtrip/2
	if clockDrift > c.clockDriftWarning || -clockDrift > c.clockDriftWarning {
		c.logger.Warnf("Estimated clock drift: %s.", clockDrift)
	}

	md, err := agentpb.GetAgentServerMetadata(stream)
	if err != nil {
		c.logger.Warnf("Can't get metadata from server: %v", err)
	}

	c.mdReader.ReadMetadata(md)

	go c.handleAppInterrupt()
	c.wg.Add(1)
	go c.handleChanges()
	c.wg.Add(1)
	go c.handleRequests()
}

func (c *Client) handleAppInterrupt() {
	select {
	case <-c.appCtx.Done():
		c.Stop()
	case <-c.done:
		return
	}
}

// Stop closes connection with server.
func (c *Client) Stop() {
	prometheus.Unregister(c.channel)
	err := c.conn.Close()
	switch err {
	case nil:
		c.logger.Info("Connection closed.")
	default:
		c.logger.Errorf("Connection closed: %s.", err)
	}
	c.wg.Wait()
	close(c.done)
}

func (c *Client) handleRequests() {
	defer c.wg.Done()

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
			c.stateCgr.SetState(payload.SetState)

			agentMessage = &agentpb.AgentMessage{
				Id: serverMessage.Id,
				Payload: &agentpb.AgentMessage_SetState{
					SetState: new(agentpb.SetStateResponse),
				},
			}

		default:
			c.logger.Panicf("Unhandled server message payload: %s.", payload)
		}

		c.channel.SendResponse(agentMessage)
	}
}

func (c *Client) handleChanges() {
	defer c.wg.Done()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for state := range c.stateCgr.Changes() {
			res := c.channel.SendRequest(&agentpb.AgentMessage_StateChanged{
				StateChanged: &state,
			})
			if res == nil {
				c.logger.Warn("Failed to send StateChanged request.")
			}
		}
		c.logger.Info("Supervisor changes done.")
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for collect := range c.stateCgr.QANRequests() {
			res := c.channel.SendRequest(&agentpb.AgentMessage_QanCollect{
				QanCollect: &collect,
			})
			if res == nil {
				c.logger.Warn("Failed to send QanCollect request.")
			}
		}
		c.logger.Info("Supervisor QAN requests done.")
	}()
}
