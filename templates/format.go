package templates

import (
	"bytes"
	"text/tabwriter"
)

// Format data.
func Format(format string, data interface{}) string {
	b := &bytes.Buffer{}
	w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)

	tmpl, err := Parse(format)
	if err != nil {
		return err.Error()
	}
	if err := tmpl.Execute(w, data); err != nil {
		return err.Error()
	}

	w.Flush()

	return b.String()
}
