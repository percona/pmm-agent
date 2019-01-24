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

package main

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/percona/pmm-agent/agentlocal"
	"github.com/percona/pmm-agent/config"
	"github.com/percona/pmm-agent/server"
	"github.com/percona/pmm-agent/supervisor"
	"github.com/percona/pmm-agent/utils/logger"
)

var (
	Version = "2.0.0-dev"
)

const (
	dialTimeout     = 10 * time.Second
	backoffMaxDelay = 10 * time.Second
)

func workLoop(ctx context.Context, cfg *config.Config, l *logrus.Entry, client agent.AgentClient) {
	// use separate context for stream to cancel it after supervisor is done sending last changes
	streamCtx, streamCancel := context.WithCancel(context.Background())
	streamCtx = agent.AddAgentConnectMetadata(streamCtx, &agent.AgentConnectMetadata{
		ID:      cfg.ID,
		Version: Version,
	})

	l.Info("Establishing two-way communication channel ...")
	stream, err := client.Connect(streamCtx)
	if err != nil {
		l.Errorf("Failed to establish two-way communication channel: %s.", err)
		streamCancel()
		return
	}

	channel := server.NewChannel(stream)
	prometheus.MustRegister(channel)
	defer func() {
		err = channel.Wait()
		switch err {
		case nil:
			l.Info("Two-way communication channel closed.")
		default:
			l.Errorf("Two-way communication channel closed: %s", err)
		}
	}()

	// So far nginx can handle all that itself without pmm-managed.
	// We need to send ping to ensure that pmm-managed is alive.
	start := time.Now()
	res := channel.SendRequest(&agent.AgentMessage_Ping{
		Ping: new(agent.Ping),
	})
	if res == nil {
		// error will be logged by channel code
		streamCancel()
		return
	}
	latency := time.Since(start) / 2
	t, err := ptypes.Timestamp(res.(*agent.ServerMessage_Pong).Pong.CurrentTime)
	if err != nil {
		l.Errorf("Failed to decode Pong.current_time: %s.", err)
		streamCancel()
		return
	}
	clockDrift := t.Sub(start.Add(latency))
	l.Infof("Two-way communication channel established. Latency: %s. Clock drift: %s.", latency, clockDrift)

	s := supervisor.NewSupervisor(ctx, &cfg.Paths, &cfg.Ports)
	go func() {
		for state := range s.Changes() {
			res := channel.SendRequest(&agent.AgentMessage_StateChanged{
				StateChanged: &state,
			})
			if res == nil {
				l.Warn("Failed to send StateChanged request.")
			}
		}
		l.Info("Supervisor done.")
		streamCancel()
	}()

	for serverMessage := range channel.Requests() {
		var agentMessage *agent.AgentMessage
		switch payload := serverMessage.Payload.(type) {
		case *agent.ServerMessage_Ping:
			agentMessage = &agent.AgentMessage{
				Id: serverMessage.Id,
				Payload: &agent.AgentMessage_Pong{
					Pong: &agent.Pong{
						CurrentTime: ptypes.TimestampNow(),
					},
				},
			}

		case *agent.ServerMessage_SetState:
			s.SetState(payload.SetState.AgentProcesses)

			agentMessage = &agent.AgentMessage{
				Id: serverMessage.Id,
				Payload: &agent.AgentMessage_SetState{
					SetState: &agent.SetStateResponse{},
				},
			}

		default:
			l.Panicf("Unhandled server message payload: %s.", payload)
		}

		channel.SendResponse(agentMessage)
	}
}

func main() {
	var cfg config.Config
	app := config.Application(&cfg, Version)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	cfg.Paths.Lookup()
	logrus.Infof("Loaded configuration: %+v.", cfg)

	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		grpclog.SetLoggerV2(&logger.GRPC{Entry: logrus.WithField("component", "gRPC")})
		logrus.Debug("Debug logging enabled.")
	}

	// TODO
	_ = agentlocal.AgentLocalServer{}

	ctx, cancel := context.WithCancel(context.Background())

	// handle termination signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, unix.SIGTERM, unix.SIGINT)
	go func() {
		s := <-signals
		signal.Stop(signals)
		logrus.Warnf("Got %s, shutting down...", unix.SignalName(s.(unix.Signal)))
		cancel()
	}()

	if cfg.Address == "" {
		logrus.Error("PMM Server address is not provided, halting.")
		<-ctx.Done()
		return
	}

	host, _, _ := net.SplitHostPort(cfg.Address)
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: cfg.InsecureTLS, //nolint:gosec
	}
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithWaitForHandshake(),
		grpc.WithBackoffMaxDelay(backoffMaxDelay),
		grpc.WithUserAgent("pmm-agent/" + Version),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}

	l := logrus.WithField("component", "connection")
	l.Infof("Connecting to %s ...", cfg.Address)
	dialCtx, dialCancel := context.WithTimeout(ctx, dialTimeout)
	conn, err := grpc.DialContext(dialCtx, cfg.Address, opts...)
	dialCancel()
	if err != nil {
		l.Fatalf("Failed to connect to %s: %s.", cfg.Address, err)
	}
	defer func() {
		err := conn.Close()
		switch err {
		case nil:
			l.Info("Connection closed.")
		default:
			l.Errorf("Connection closed: %s.", err)
		}
	}()

	l.Infof("Connected to %s.", cfg.Address)
	client := agent.NewAgentClient(conn)

	// TODO
	// if cfg.UUID == "" {
	// 	logrus.Info("Registering pmm-agent ...")
	// 	resp, err := client.Register(ctx, &agent.RegisterRequest{})
	// 	if err != nil {
	// 		logrus.Fatalf("Failed to register pmm-agent: %s.", err)
	// 	}
	// 	cfg.UUID = resp.Uuid
	// 	logrus.Infof("pmm-agent registered: %s.", cfg.UUID)
	// }

	workLoop(ctx, &cfg, l, client)
}
