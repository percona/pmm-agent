package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/app"
	"github.com/percona/pmm-agent/cmd/add"
	"github.com/percona/pmm-agent/cmd/list"
	"github.com/percona/pmm-agent/cmd/remove"
	"github.com/percona/pmm-agent/cmd/root"
	"github.com/percona/pmm-agent/cmd/serve"
	"github.com/percona/pmm-agent/cmd/start"
	"github.com/percona/pmm-agent/cmd/stop"
)

// New returns app cmd.
func New(app *app.App) *cobra.Command {
	cmd := root.New(app)

	cmd.AddCommand(
		serve.New(context.Background(), app),
		add.New(app),
		remove.New(app),
		start.New(app),
		stop.New(app),
		list.New(app),
	)

	return cmd
}
