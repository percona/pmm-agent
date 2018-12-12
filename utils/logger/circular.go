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

package logger

import (
	"bytes"
	"sync"
)

// CircularWriter is a Writer that holds several latest lines written.
type CircularWriter struct {
	m    sync.RWMutex
	buf  []byte
	i    int
	data []string
}

// New creates new CircularWriter with a given amount of lines to hold.
func New(lines int) *CircularWriter {
	writer := CircularWriter{
		data: make([]string, lines),
	}
	return &writer
}

// Write splits lines and add them to lines list.
func (c *CircularWriter) Write(p []byte) (n int, err error) {
	b := bytes.NewBuffer(c.buf)
	n, err = b.Write(p)
	if err != nil {
		return
	}

	c.m.Lock()
	defer c.m.Unlock()

	var line string
	for {
		line, err = b.ReadString('\n')
		if err != nil {
			c.buf = []byte(line)
			err = nil
			return
		}
		c.data[c.i] = line[:len(line)-1]
		c.i = (c.i + 1) % len(c.data)
	}
}

// Data returns string array from circular list.
func (c *CircularWriter) Data() []string {
	c.m.RLock()
	defer c.m.RUnlock()

	var result []string
	for i := c.i; i < c.i+len(c.data); i++ {
		data := c.data[i%len(c.data)]
		if data != "" {
			result = append(result, data)
		}
	}
	return result
}
