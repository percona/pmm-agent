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
	"os"
	"os/signal"
	"sync"

	"github.com/percona/pmm/version"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/grpclog"

	"github.com/percona/pmm-agent/agentlocal"
	"github.com/percona/pmm-agent/agents/supervisor"
	"github.com/percona/pmm-agent/client"
	"github.com/percona/pmm-agent/config"
	"github.com/percona/pmm-agent/utils/logger"
)

func main() {
	// empty version breaks much of pmm-managed logic
	if version.Version == "" {
		panic("pmm-agent version is not set during build.")
	}

	l := logrus.WithField("component", "main")
	ctx, cancel := context.WithCancel(context.Background())
	defer l.Info("Done.")

	// handle termination signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, unix.SIGTERM, unix.SIGINT)
	go func() {
		s := <-signals
		signal.Stop(signals)
		logrus.Warnf("Got %s, shutting down...", unix.SignalName(s.(unix.Signal)))
		cancel()
	}()

	var grpclogOnce sync.Once
	for ctx.Err() == nil {
		cfg, err := config.Parse(logrus.WithField("component", "config"))
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Debugf("Loaded configuration: %+v", cfg)

		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetReportCaller(false)
		if cfg.Debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if cfg.Trace {
			logrus.SetLevel(logrus.TraceLevel)
			logrus.SetReportCaller(true)
		}
		grpclogOnce.Do(func() {
			if cfg.Trace {
				grpclog.SetLoggerV2(&logger.GRPC{Entry: logrus.WithField("component", "grpclog")})
			}
		})

		for ctx.Err() == nil {
			appCtx, appCancel := context.WithCancel(ctx)

			supervisor := supervisor.NewSupervisor(appCtx, &cfg.Paths, &cfg.Ports)
			localServer := agentlocal.NewServer(cfg, supervisor)
			client := client.New(cfg, supervisor, localServer)

			server := make(chan error)
			go func() {
				server <- localServer.Run(appCtx, client)
			}()

			// FIXME
			// Deadlock when pmm-agent is started without ID and /local/Reload is called.

			_ = client.Run(appCtx)
			appCancel()
			<-client.Done()

			err = <-server
			if err == agentlocal.ErrReload {
				break
			}
		}
	}
}
