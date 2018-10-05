package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/api"
	"github.com/percona/pmm-agent/errs"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop PROGRAM",
	Short: "Stop PROGRAM.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var errs errs.Errs
		if len(args) == 0 {
			req := &api.StopAllRequest{}
			return client.Call(func(ctx context.Context, c api.SupervisorClient) error {
				_, err := c.StopAll(ctx, req)
				return err
			})
		}
		for i := range args {
			req := &api.StopRequest{
				Name: args[i],
			}
			err := client.Call(func(ctx context.Context, c api.SupervisorClient) error {
				_, err := c.Stop(ctx, req)
				return err
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errs
		}
		return nil
	},
}

func init() {
	client.Flags(stopCmd)
	rootCmd.AddCommand(stopCmd)
}
