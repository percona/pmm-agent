package parser

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTables(t *testing.T) {
	files, err := filepath.Glob(filepath.FromSlash("./testdata/*.sql"))
	require.NoError(t, err)
	for _, file := range files {
		goldenFile := strings.TrimSuffix(file, ".sql") + ".json"
		name := strings.TrimSuffix(filepath.Base(file), ".log")
		t.Run(name, func(t *testing.T) {
			b, err := ioutil.ReadFile(file) //nolint:gosec
			require.NoError(t, err)
			query := string(b)
			actual, err := ExtractTables(query)
			require.NoError(t, err)

			b, err = ioutil.ReadFile(goldenFile) //nolint:gosec
			require.NoError(t, err)
			var expected []string
			err = json.Unmarshal(b, &expected)
			require.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	}
}
