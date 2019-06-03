// pmm-agent
// Copyright (C) 2018 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package parser

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanup(t *testing.T, files []string) {
	t.Helper()
	for _, f := range files {
		assert.NoError(t, os.Remove(f))
	}
}

func TestContinuousFileReader(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		f, err := ioutil.TempFile("", "pmm-agent-test-slowlog-reader")
		require.NoError(t, err)
		files := []string{f.Name()}
		defer func() {
			cleanup(t, files)
		}()

		_, err = f.WriteString("0\n")
		require.NoError(t, err)

		r, err := NewContinuousFileReader(f.Name(), &testLogger{t})
		require.NoError(t, err)
		r.sleep = 50 * time.Millisecond
		assert.Equal(t, &ReaderMetrics{InputSize: 2, InputPos: 2}, r.Metrics())

		done := make(chan struct{})
		defer func() {
			_ = r.Close()
			<-done
		}()

		// use separate goroutine to give race detector more chances to spot problems
		lines := make(chan string, 10)
		go func() {
			for {
				line, e := r.NextLine()
				if e != nil {
					assert.Equal(t, io.EOF, e, "%s", e)
					close(lines)
					close(done)
					return
				}
				lines <- line
			}
		}()

		// test partial line read
		_, err = f.WriteString("1\n2\n3")
		require.NoError(t, err)
		assert.Equal(t, "1\n", <-lines)
		assert.Equal(t, "2\n", <-lines)
		assert.Empty(t, lines)
		assert.Equal(t, &ReaderMetrics{InputSize: 7, InputPos: 7}, r.Metrics())

		// test rotation with rename
		newName := f.Name() + "-1"
		require.NoError(t, os.Rename(f.Name(), newName))
		files = append(files, newName)
		f, err = os.Create(f.Name())
		require.NoError(t, err)
		_, err = f.WriteString("\n4\n5")
		require.NoError(t, err)
		assert.Equal(t, "3\n", <-lines)
		assert.Equal(t, "4\n", <-lines)
		assert.Empty(t, lines)
		assert.Equal(t, &ReaderMetrics{InputSize: 4, InputPos: 4}, r.Metrics())

		// test rotation with truncate
		f, err = os.Create(f.Name())
		require.NoError(t, err)
		_, err = f.WriteString("\n6\n") // new file should be smaller
		require.NoError(t, err)
		assert.Equal(t, "5\n", <-lines)
		assert.Equal(t, "6\n", <-lines)
		assert.Empty(t, lines)
		assert.Equal(t, &ReaderMetrics{InputSize: 3, InputPos: 3}, r.Metrics())

		// test close
		_, err = f.WriteString("7\n8")
		require.NoError(t, err)
		assert.NoError(t, r.Close())
		l, ok := <-lines
		assert.False(t, ok, "line = %q", l)
		assert.Nil(t, r.Metrics())
	})
}
