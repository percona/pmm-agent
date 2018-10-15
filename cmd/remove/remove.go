package remove

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/api"
	"github.com/percona/pmm-agent/app"
	"github.com/percona/pmm-agent/errs"
)

// New returns `remove` command.
func New(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove PROGRAM",
		Short: "Stop and remove PROGRAM.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var errs errs.Errs
			if len(args) == 0 {
				req := &api.RemoveAllRequest{}
				return app.Client.Call(func(ctx context.Context, c api.SupervisorClient) error {
					_, err := c.RemoveAll(ctx, req)
					return err
				})
			}
			for i := range args {
				req := &api.RemoveRequest{
					Name: args[i],
				}
				err := app.Client.Call(func(ctx context.Context, c api.SupervisorClient) error {
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

	app.Client.Flags(cmd)
	return cmd
}
