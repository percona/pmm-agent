package client

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/percona/pmm-agent/api"
)

// CallFunc allows to execute actions on ready to use SupervisorClient.
type CallFunc = func(ctx context.Context, client api.SupervisorClient) error

// Client contains configuration required to call a server.
type Client struct {
	Timeout time.Duration
	Addr    string
}

// Call server and execute provided function f (CallFunc).
func (c Client) Call(f CallFunc) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.Addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	return f(ctx, api.NewSupervisorClient(conn))
}

// Flags assigns flags to cmd.
func (c *Client) Flags(cmd *cobra.Command) {
	cmd.Flags().DurationVar(&c.Timeout, "timeout", 10*time.Second, "timeout")
	cmd.Flags().StringVar(&c.Addr, "addr", "127.0.0.1:7771", "gRPC server address")
}
