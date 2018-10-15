package add

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/api"
	"github.com/percona/pmm-agent/app"
)

// New returns `add` command.
func New(app *app.App) *cobra.Command {
	var env []string
	cmd := &cobra.Command{
		Use:   "add PROGRAM [--env VAR=VALUE] -- NAME [ARGUMENTS]",
		Short: "Add PROGRAM with given NAME, ARGUMENTS and set of environment variables provided with --env.",
		Args: func(cmd *cobra.Command, args []string) error {
			switch cmd.ArgsLenAtDash() {
			case -1:
				return fmt.Errorf("missing double dash '--'")
			case 0:
				return fmt.Errorf("missing program name")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &api.AddRequest{
				Name: args[0],
				Program: &api.Program{
					Name: args[1],
					Arg:  args[2:],
					Env:  env,
				},
			}

			return app.Client.Call(func(ctx context.Context, c api.SupervisorClient) error {
				_, err := c.Add(ctx, req)
				return err
			})
		},
	}

	app.Client.Flags(cmd)
	cmd.Flags().StringArrayVar(&env, "env", nil, "environment variable")
	return cmd
}
