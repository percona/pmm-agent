package tests

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"gopkg.in/reform.v1"
)

func LogTable(t *testing.T, structs []reform.Struct) {
	t.Helper()

	if len(structs) == 0 {
		t.Log("logTable: empty")
		return
	}

	columns := structs[0].View().Columns()
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.Debug)
	_, err := fmt.Fprintln(w, strings.Join(columns, "\t"))
	require.NoError(t, err)
	for i, c := range columns {
		columns[i] = strings.Repeat("-", len(c))
	}
	_, err = fmt.Fprintln(w, strings.Join(columns, "\t"))
	require.NoError(t, err)

	for _, str := range structs {
		res := make([]string, len(str.Values()))
		for i, v := range str.Values() {
			res[i] = spew.Sprint(v)
		}
		fmt.Fprintf(w, "%s\n", strings.Join(res, "\t"))
	}

	require.NoError(t, w.Flush())
	t.Logf("%s:\n%s", structs[0].View().Name(), buf.Bytes())
}
