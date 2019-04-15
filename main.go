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
	_ "expvar"         // register /debug/vars
	_ "net/http/pprof" // register /debug/pprof
	"os"
	"os/signal"

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

	cfg, err := config.Get(os.Args[1:], logrus.WithField("component", "config"))
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("Loaded configuration: %+v", cfg)

	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if cfg.Trace {
		logrus.SetLevel(logrus.TraceLevel)
		logrus.SetReportCaller(true)
		grpclog.SetLoggerV2(&logger.GRPC{Entry: logrus.WithField("component", "grpclog")})
	}

	appContext, shutdown := context.WithCancel(context.Background())
	reloadSignal := make(chan bool)

	handleTerminationSignals(shutdown)

	if cfg.Address == "" {
		logrus.Error("PMM Server address is not provided, halting.")
		<-appContext.Done()
	}

	svr := supervisor.NewSupervisor(appContext, &cfg.Paths, &cfg.Ports)
	srv := agentlocal.NewServer(appContext, cfg, reloadSignal)
	clt := client.New(appContext, cfg, svr, srv)

	r := &reloader{
		appCtx:       appContext,
		reloadSignal: reloadSignal,
		server:       srv,
		client:       clt,
		supervisor:   svr,
	}
	r.Watch()

	srv.Run()
	clt.Run()

	// FIXME svr.Wait()
	srv.Wait()
	clt.Wait()
}

type reloader struct {
	appCtx       context.Context
	reloadSignal chan bool
	server       *agentlocal.Server
	client       *client.Client
	supervisor   *supervisor.Supervisor
}

// Watch runs goroutine for which watches reloadSignal and performs components reload.
func (r *reloader) Watch() {
	go func() {
		for {
			select {
			case <-r.appCtx.Done():
				return
			case <-r.reloadSignal:
				r.reload()
			}
		}
	}()
}

func (r *reloader) reload() {
	logrus.Warnf("Got restart signal, restarting...")

	cfg, err := config.Get(os.Args[1:], logrus.WithField("component", "config"))
	if err != nil {
		logrus.Fatal(err)
	}

	// FIXME
	// r.supervisor.Stop()
	// r.supervisor.Wait()

	r.client.Stop()
	r.client.Wait()

	r.supervisor = supervisor.NewSupervisor(r.appCtx, &cfg.Paths, &cfg.Ports)
	r.client = client.New(r.appCtx, cfg, r.supervisor, r.server)

	r.client.Run()
}

func handleTerminationSignals(shutdown context.CancelFunc) {
	// handle termination signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, unix.SIGTERM, unix.SIGINT)
	go func() {
		s := <-signals
		signal.Stop(signals)
		logrus.Warnf("Got %s, shutting down...", unix.SignalName(s.(unix.Signal)))
		shutdown()
	}()
}
