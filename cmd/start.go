package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/api"
	"github.com/percona/pmm-agent/errs"
)

// startCmd represents the stop command
var startCmd = &cobra.Command{
	Use:   "start PROGRAM",
	Short: "Start PROGRAM.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var errs errs.Errs
		if len(args) == 0 {
			req := &api.StartAllRequest{}
			return client.Call(func(ctx context.Context, c api.SupervisorClient) error {
				_, err := c.StartAll(ctx, req)
				return err
			})
		}
		for i := range args {
			req := &api.StartRequest{
				Name: args[i],
			}
			err := client.Call(func(ctx context.Context, c api.SupervisorClient) error {
				_, err := c.Start(ctx, req)
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
	client.Flags(startCmd)
	rootCmd.AddCommand(startCmd)
}
