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

package agentlocal

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/percona/pmm/api/agentlocalpb"
	"github.com/percona/pmm/api/agentpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	channelz "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"

	"github.com/percona/pmm-agent/config"
)

type agentsGetter interface {
	AgentsList() []*agentlocalpb.AgentInfo
}

const (
	shutdownTimeout = 3 * time.Second
	defaultGrpcAddr = "127.0.0.1:7776"
	defaultJSONAddr = "127.0.0.1:7777"
)

// Server represents local agent api server.
type Server struct {
	gRPCAddress string
	gRPCTimeout time.Duration
	cfg         *config.Config
	gRPCServer  *grpc.Server
	jsonServer  *http.Server
	jsonLogger  logrus.FieldLogger
	grpcLogger  logrus.FieldLogger
	appCtx      context.Context

	rw                 sync.RWMutex
	ag                 agentsGetter
	currentSrvMetadata *agentpb.AgentServerMetadata

	reload chan<- bool
	done   chan bool

	wg sync.WaitGroup
}

// NewServer creates new local agent api server instance.
func NewServer(appCtx context.Context, ag agentsGetter, cfg *config.Config, reload chan<- bool) *Server {
	instance := &Server{
		appCtx:      appCtx,
		gRPCAddress: defaultGrpcAddr,
		gRPCTimeout: shutdownTimeout,
		cfg:         cfg,
		jsonLogger:  logrus.WithField("component", "JSON"),
		grpcLogger:  logrus.WithField("component", "gRPC"),
		ag:          ag,
		reload:      reload,
		done:        make(chan bool),
	}

	return instance
}

// Wait blocks until server end its work.
func (s *Server) Wait() {
	<-s.done
}

// ReadMetadata reads new metadata from provided metadata structure to local state.
func (s *Server) ReadMetadata(md agentpb.AgentServerMetadata) {
	s.rw.Lock()
	s.currentSrvMetadata = &md
	s.rw.Unlock()
}

// Status returns local agent status.
func (s *Server) Status(ctx context.Context, req *agentlocalpb.StatusRequest) (*agentlocalpb.StatusResponse, error) {
	md := s.getMetadata()

	var user *url.Userinfo
	switch {
	case s.cfg.Password != "":
		user = url.UserPassword(s.cfg.Username, s.cfg.Password)
	case s.cfg.Username != "":
		user = url.User(s.cfg.Username)
	}
	u := url.URL{
		Scheme: "https",
		User:   user,
		Host:   s.cfg.Address,
		Path:   "/",
	}
	srvInfo := &agentlocalpb.ServerInfo{
		Url:          u.String(),
		InsecureTls:  s.cfg.InsecureTLS,
		Version:      md.ServerVersion,
		LastPingTime: nil, // TODO https://jira.percona.com/browse/PMM-3758
		Latency:      nil, // TODO https://jira.percona.com/browse/PMM-3758
	}

	agentsInfo := s.ag.AgentsList()

	return &agentlocalpb.StatusResponse{
		AgentId:      s.cfg.ID,
		RunsOnNodeId: md.AgentRunsOnNodeID,
		ServerInfo:   srvInfo,
		AgentsInfo:   agentsInfo,
	}, nil
}

// Reload reloads pmm-agent and it configuration.
func (s *Server) Reload(ctx context.Context, req *agentlocalpb.ReloadRequest) (*agentlocalpb.ReloadResponse, error) {
	defer func() {
		s.reload <- true
	}()
	return &agentlocalpb.ReloadResponse{}, nil
}

// Run runs gRPC server with JSON proxy until context is canceled, then gracefully stops it..
func (s *Server) Run() {
	go s.handleAppInterrupt()
	s.wg.Add(1)
	go s.runGRPCServer()
	s.wg.Add(1)
	go s.runJSONServer()
}

// runGRPCServer runs gRPC server until context is canceled, then gracefully stops it.
func (s *Server) runGRPCServer() {
	defer s.wg.Done()

	s.grpcLogger.Infof("Starting server on http://%s/ ...", s.gRPCAddress)

	s.gRPCServer = grpc.NewServer()
	agentlocalpb.RegisterAgentLocalServer(s.gRPCServer, s)

	if s.cfg.Debug {
		s.grpcLogger.Debug("Reflection and channelz are enabled.")
		reflection.Register(s.gRPCServer)
		channelz.RegisterChannelzServiceToServer(s.gRPCServer)
	}

	// run server until it is stopped gracefully or not
	listener, err := net.Listen("tcp", s.gRPCAddress)
	if err != nil {
		s.grpcLogger.Panic(err)
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			err = s.gRPCServer.Serve(listener)
			if err == nil || err == grpc.ErrServerStopped {
				break
			}
			s.grpcLogger.Errorf("Failed to serve: %s", err)
		}
		s.grpcLogger.Info("gRPC server stopped.")
	}()
}

func (s *Server) handleAppInterrupt() {
	select {
	case <-s.appCtx.Done():
		s.Stop()
	case <-s.done:
		return
	}
}

// Stop stops server.
func (s *Server) Stop() {
	// try to stop json server gracefully, then not
	jCtx, jCancel := context.WithTimeout(context.Background(), s.gRPCTimeout)
	if err := s.jsonServer.Shutdown(jCtx); err != nil {
		s.jsonLogger.Errorf("Failed to shutdown gracefully: %s", err)
	}
	jCancel()

	// try to stop server gracefully, then not
	ctx, cancel := context.WithTimeout(context.Background(), s.gRPCTimeout)
	go func() {
		<-ctx.Done()
		s.gRPCServer.Stop()

		<-jCtx.Done()
		s.wg.Wait()
		close(s.done)
	}()
	s.gRPCServer.GracefulStop()
	cancel()
}

// runJSONServer runs JSON proxy server (grpc-gateway) until context is canceled, then gracefully stops it.
func (s *Server) runJSONServer() {
	defer s.wg.Done()

	s.jsonLogger.Infof("Starting server on http://%s/ ...", defaultJSONAddr)

	proxyMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := agentlocalpb.RegisterAgentLocalHandlerFromEndpoint(s.appCtx, proxyMux, s.gRPCAddress, opts); err != nil {
		s.jsonLogger.Panic(err)
	}

	handlers := []string{
		"/debug/vars",     // by expvar
		"/debug/requests", // by golang.org/x/net/trace imported by google.golang.org/grpc
		"/debug/events",   // by golang.org/x/net/trace imported by google.golang.org/grpc
		"/debug/pprof",    // by net/http/pprof
	}
	for i, h := range handlers {
		handlers[i] = "http://" + defaultJSONAddr + h
	}

	var buf bytes.Buffer
	err := template.Must(template.New("debug").Parse(`
	<html>
	<body>
	<ul>
	{{ range . }}
		<li><a href="{{ . }}">{{ . }}</a></li>
	{{ end }}
	</ul>
	</body>
	</html>
	`)).Execute(&buf, handlers)
	if err != nil {
		s.jsonLogger.Panic(err)
	}

	http.Handle("/", proxyMux)
	http.HandleFunc("/debug", func(rw http.ResponseWriter, req *http.Request) {
		if _, err := rw.Write(buf.Bytes()); err != nil {
			s.jsonLogger.Warn(err)
		}
	})

	s.jsonLogger.Infof("Starting server on http://%s/debug\nRegistered handlers:\n\t%s", defaultJSONAddr, strings.Join(handlers, "\n\t"))

	s.jsonServer = &http.Server{
		Addr:     defaultJSONAddr,
		ErrorLog: log.New(os.Stderr, "runJSONServer: ", 0),
		Handler:  http.DefaultServeMux,
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.jsonServer.ListenAndServe(); err != http.ErrServerClosed {
			s.jsonLogger.Panic(err)
		}
		s.jsonLogger.Info("JSON server stopped.")
	}()
}

func (s *Server) getMetadata() agentpb.AgentServerMetadata {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return *s.currentSrvMetadata
}

// check interfaces
var (
	_ agentlocalpb.AgentLocalServer = (*Server)(nil)
)
