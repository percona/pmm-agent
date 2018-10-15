package root

import (
	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/app"
)

// New returns root command.
func New(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pmm-agent",
		Short:         "pmm-agent manages PMM node.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	app.Config.Flags(cmd)
	return cmd
}
