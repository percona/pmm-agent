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

package supervisor

import (
	"bytes"
	"sync"

	"github.com/AlekSi/pointer"
)

// circularWriter is a Writer that holds several latest lines written.
type circularWriter struct {
	m    sync.RWMutex
	buf  []byte
	i    int
	data []*string
}

// newCircularWriter creates new circularWriter with a given amount of lines to hold.
func newCircularWriter(lines int) *circularWriter {
	return &circularWriter{
		data: make([]*string, lines),
	}
}

// Write splits lines and add them to lines list.
// This method is thread-safe.
func (cw *circularWriter) Write(p []byte) (n int, err error) {
	cw.m.Lock()
	defer cw.m.Unlock()

	b := bytes.NewBuffer(cw.buf)
	n, err = b.Write(p)
	if err != nil {
		return
	}

	var line string
	for {
		line, err = b.ReadString('\n')
		if err != nil {
			cw.buf = []byte(line)
			err = nil
			return
		}
		cw.data[cw.i] = pointer.ToString(line[:len(line)-1])
		cw.i = (cw.i + 1) % len(cw.data)
	}
}

// Data returns string array from circular list.
// This method is thread-safe.
func (cw *circularWriter) Data() []string {
	cw.m.RLock()
	defer cw.m.RUnlock()

	result := make([]string, 0, len(cw.data))
	for i := cw.i; i < cw.i+len(cw.data); i++ {
		line := cw.data[i%len(cw.data)]
		if line != nil {
			result = append(result, *line)
		}
	}
	return result
}
