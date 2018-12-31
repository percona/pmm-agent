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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
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

func workLoop(ctx context.Context, cfg *config.Config, client agent.AgentClient) {
	l := logrus.WithField("component", "conn")

	l.Info("Establishing two-way communication channel ...")
	ctx = agent.AddAgentConnectMetadata(ctx, &agent.AgentConnectMetadata{
		UUID:    cfg.UUID,
		Version: Version,
	})
	stream, err := client.Connect(ctx)
	if err != nil {
		l.Fatal(err)
	}
	l.Info("Two-way communication channel established.")

	channel := server.NewChannel(stream)
	prometheus.MustRegister(channel)

	svr := supervisor.NewSupervisor(ctx, cfg.Ports)
	go func() {
		for {
			select {
			case update, more := <-svr.StateUpdates():
				l.Debugf("Agent %d changed state to %s", update.AgentID, update.State)
				if !more {
					return
				}
			}
		}
	}()

	for serverMessage := range channel.Requests() {
		var agentMessage *agent.AgentMessage
		switch payload := serverMessage.Payload.(type) {
		case *agent.ServerMessage_Ping:
			agentMessage = &agent.AgentMessage{
				Id: serverMessage.Id,
				Payload: &agent.AgentMessage_Ping{
					Ping: &agent.PingResponse{
						CurrentTime: ptypes.TimestampNow(),
					},
				},
			}

		case *agent.ServerMessage_State:
			agentProcessesStates := svr.UpdateState(payload.State.AgentProcesses)

			agentMessage = &agent.AgentMessage{
				Id: serverMessage.Id,
				Payload: &agent.AgentMessage_State{
					State: &agent.SetStateResponse{
						AgentProcesses: agentProcessesStates,
					},
				},
			}

		default:
			l.Panicf("Unhandled server message payload: %s.", payload)
		}

		channel.SendResponse(agentMessage)
	}

	l.Error(channel.Wait())
}

func main() {
	var cfg config.Config
	app := config.Application(&cfg, Version)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	logrus.Infof("Loaded configuration: %+v.", cfg)

	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		grpclog.SetLoggerV2(&logger.GRPC{Entry: logrus.WithField("component", "gRPC")})
		logrus.Debug("Debug logging enabled.")
	}

	_ = agentlocal.AgentLocalServer{}

	ctx, cancelFunc := context.WithCancel(context.Background())

	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		logrus.Debugf("caught sig: %+v", sig)
		cancelFunc()
	}()

	if cfg.Address == "" {
		logrus.Error("PMM Server address is not provided, halting.")
		<-ctx.Done()
		return
	}

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithWaitForHandshake(),
		grpc.WithBackoffMaxDelay(backoffMaxDelay),
		grpc.WithUserAgent("pmm-agent/" + Version),
	}
	if cfg.WithoutNginx {
		opts = append(opts, grpc.WithInsecure())
	} else {
		creds := credentials.NewTLS(&tls.Config{})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	logrus.Infof("Connecting to %s ...", cfg.Address)
	dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	conn, err := grpc.DialContext(dialCtx, cfg.Address, opts...)
	cancel()
	if err != nil {
		logrus.Fatalf("Failed to connect to %s: %s.", cfg.Address, err)
	}
	logrus.Infof("Connected to %s.", cfg.Address)
	client := agent.NewAgentClient(conn)

	// if cfg.UUID == "" {
	// 	logrus.Info("Registering pmm-agent ...")
	// 	resp, err := client.Register(ctx, &agent.RegisterRequest{})
	// 	if err != nil {
	// 		logrus.Fatalf("Failed to register pmm-agent: %s.", err)
	// 	}
	// 	cfg.UUID = resp.Uuid
	// 	logrus.Infof("pmm-agent registered: %s.", cfg.UUID)
	// }

	workLoop(ctx, &cfg, client)
}
