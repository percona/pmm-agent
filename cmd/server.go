package cmd

import (
	"context"
	"log"
	"net"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/percona/pmm-agent/api"
	"github.com/percona/pmm-agent/errs"
	"github.com/percona/pmm-agent/handlers"
	"github.com/percona/pmm-agent/supervisor"
)

// Server contains configuration required to start a server.
type Server struct {
	Addr   string
	LogDir string
}

// Serve accepts incoming connections until ctx is canceled.
func (s Server) Serve(ctx context.Context) error {
	var errs errs.Safe
	supervisor := &supervisor.Supervisor{
		LogDir: s.LogDir,
	}

	gRPCServer := grpc.NewServer()
	api.RegisterSupervisorServer(gRPCServer, &handlers.SupervisorServer{
		Supervisor: supervisor,
	})

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	done := make(chan struct{})
	go func() {
		log.Printf("Listen on: %s", listener.Addr())
		if err := gRPCServer.Serve(listener); err != nil {
			errs.Add(err)
		}
		close(done)
	}()
	<-ctx.Done()
	gRPCServer.GracefulStop()
	<-done

	if err := supervisor.StopAll(); err != nil {
		errs.Add(err)
	}

	return errs.Err()
}

// Flags assigns flags to cmd.
func (s *Server) Flags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.Addr, "addr", "127.0.0.1:7771", "gRPC server listen address")
	cmd.Flags().StringVar(&s.LogDir, "log-dir", "/var/log", "directory for log files")
}
