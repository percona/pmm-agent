package format

import (
	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/templates"
)

// Format contains configuration required to format data.
type Format struct {
	format string
	json   bool
}

// Parse format and fill it with data.
func (f Format) Parse(format string, data interface{}) string {
	if f.format != "" {
		format = f.format
	}
	if f.json {
		format = "{{ json . }}"
	}

	return templates.Parse(format, data)
}

// Flags assigns flags to cmd.
func (f *Format) Flags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.format, "template", "", "print result using a Go template")
	cmd.Flags().BoolVar(&f.json, "json", false, "print result as json")
}
