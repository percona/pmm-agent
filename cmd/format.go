package cmd

import (
	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/templates"
)

// Format contains configuration required to format data.
type Format struct {
	format string
	json   bool
}

// Format data to given format.
func (f Format) Format(format string, data interface{}) string {
	if f.format != "" {
		format = f.format
	}
	if f.json {
		format = "{{ json . }}"
	}

	return templates.Format(format, data)
}

// Flags assigns flags to cmd.
func (f *Format) Flags(cmd *cobra.Command) {
	listCmd.Flags().StringVar(&f.format, "template", "", "print result using a Go template")
	listCmd.Flags().BoolVar(&f.json, "json", false, "print result as json")
}
