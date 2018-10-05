package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/percona/pmm-agent/api"
)

type CallFunc = func(ctx context.Context, client api.SupervisorClient) error

type Client struct {
	Timeout time.Duration
	Addr    string
}

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

func (c *Client) Flags(cmd *cobra.Command) {
	cmd.Flags().DurationVar(&c.Timeout, "timeout", 10*time.Second, "timeout")
	cmd.Flags().StringVar(&c.Addr, "addr", "127.0.0.1:7771", "gRPC server address")
}
