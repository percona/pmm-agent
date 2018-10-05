package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/api"
	"github.com/percona/pmm-agent/errs"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove PROGRAM",
	Short: "Stop and remove PROGRAM.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var errs errs.Errs
		if len(args) == 0 {
			req := &api.RemoveAllRequest{}
			return client.Call(func(ctx context.Context, c api.SupervisorClient) error {
				_, err := c.RemoveAll(ctx, req)
				return err
			})
		}
		for i := range args {
			req := &api.RemoveRequest{
				Name: args[i],
			}
			err := client.Call(func(ctx context.Context, c api.SupervisorClient) error {
				_, err := c.Remove(ctx, req)
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
	client.Flags(removeCmd)
	rootCmd.AddCommand(removeCmd)
}
