package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/api"
)

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List programs.",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return client.Call(func(ctx context.Context, c api.SupervisorClient) error {
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
			}
			cmd.Print(l)
			return nil
		})
	},
}

func init() {
	client.Flags(listCmd)
	format.Flags(listCmd)
	rootCmd.AddCommand(listCmd)
}

type list struct {
	Statuses map[string]*api.Status
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
	return format.Format(f, l)
}
