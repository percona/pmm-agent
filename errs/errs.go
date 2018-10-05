package errs

import (
	"bytes"
	"fmt"
)

// Ensure it matches error interface.
var _ error = Errs{}

// Errs is a slice of errors implementing the error interface.
type Errs []error

// Error
func (errs Errs) Error() string {
	if len(errs) == 0 {
		return ""
	}
	if len(errs) == 1 {
		return errs[0].Error()
	}
	buf := &bytes.Buffer{}
	for _, err := range errs {
		fmt.Fprintf(buf, "\n* %s", err)
	}
	return buf.String()
}
