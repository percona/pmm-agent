package cmd

import (
	"github.com/spf13/cobra"

	"github.com/percona/pmm-agent/templates"
)

type Format struct {
	format string
	json   bool
}

func (f Format) Format(format string, data interface{}) string {
	if f.format != "" {
		format = f.format
	}
	if f.json {
		format = "{{ json . }}"
	}

	return templates.Format(format, data)
}

func (f *Format) Flags(cmd *cobra.Command) {
	listCmd.Flags().StringVar(&f.format, "template", "", "print result using a Go template")
	listCmd.Flags().BoolVar(&f.json, "json", false, "print result as json")
}
