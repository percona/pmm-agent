package list

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/api"
	"github.com/percona/pmm-agent/app"
	"github.com/percona/pmm-agent/app/format"
)

// New returns `list` command.
func New(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List programs.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Client.Call(func(ctx context.Context, c api.SupervisorClient) error {
				req := &api.ListRequest{}
				resp, err := c.List(ctx, req)
				if err != nil {
					return err
				}
				if resp.Statuses == nil {
					return nil
				}
				l := &list{
					Statuses: resp.Statuses,
					Format:   &app.Format,
				}
				cmd.Print(l)
				return nil
			})
		},
	}

	app.Client.Flags(cmd)
	app.Format.Flags(cmd)
	return cmd
}

type list struct {
	Statuses map[string]*api.Status
	Format   *format.Format
}

// String representation of the list.
func (l *list) String() string {
	header := []string{
		"Name",
		"Program",
		"Arg",
		"Env",
		"Running",
		"PID",
		"Err",
	}
	field := []string{
		"{{$index}}",
		"{{.Program.Name}}",
		"{{.Program.Arg}}",
		"{{.Program.Env}}",
		"{{.Running}}",
		"{{.Pid}}",
		"{{.Err}}",
	}
	headers := strings.Join(header, "\t")
	fields := strings.Join(field, "\t")
	f := headers + "\n{{range $index, $element := .Statuses}}" + fields + "\n{{end}}"
	return l.Format.Parse(f, l)
}
