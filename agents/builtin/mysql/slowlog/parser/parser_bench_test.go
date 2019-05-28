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
	"bufio"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/percona/go-mysql/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkParser(b *testing.B) {
	files, err := filepath.Glob(filepath.FromSlash("./testdata/*.log"))
	require.NoError(b, err)
	for _, name := range files {
		benchmarkFile(b, name)
	}
}

func benchmarkFile(b *testing.B, name string) {
	b.Helper()

	f, err := os.Open(name) //nolint:gosec
	require.NoError(b, err)
	defer func() {
		require.NoError(b, f.Close())
	}()

	b.Run(name, func(b *testing.B) {
		s, err := f.Stat()
		require.NoError(b, err)
		b.SetBytes(s.Size())
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			b.StopTimer()

			_, err = f.Seek(0, io.SeekStart)
			assert.NoError(b, err)
			r := bufio.NewReader(f)
			p := NewSlowLogParser(r, log.Options{})
			done := make(chan error)

			b.StartTimer()

			go func() {
				done <- p.Start()
			}()
			for range p.EventChan() {
			}
			assert.NoError(b, <-done)
		}
	})
}
