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
	"bufio"
	"sync"

	"github.com/sirupsen/logrus"
)

//CircularWriter is a Writer which write in circular list.
type CircularWriter struct {
	data []string
	i    int
	l    *logrus.Entry
	m    *sync.RWMutex
	buf  []byte
}

// New creates new CircularWriter.
func New(len int) *CircularWriter {
	writer := CircularWriter{
		data: make([]string, len),
		i:    0,
		l:    logrus.WithField("component", "CircularWriter"),
		m:    &sync.RWMutex{},
	}
	return &writer
}

// Write writes new data into buffer.
func (c *CircularWriter) Write(p []byte) (n int, err error) {
	c.m.Lock()
	defer c.m.Unlock()
	c.buf = append(c.buf, p...)
	for {
		advance, token, err := bufio.ScanLines(c.buf, false)
		if err != nil {
			return 0, err
		}
		if token == nil {
			break
		}
		c.i = (c.i + 1) % len(c.data)
		c.data[c.i] = string(token)
		c.buf = c.buf[advance:]
	}
	return len(p), nil
}

// Data returns string array from circular list.
func (c *CircularWriter) Data() []string {
	c.m.RLock()
	defer c.m.RUnlock()

	var result []string
	currentIndex := c.i
	for i := currentIndex + 1; i <= currentIndex+len(c.data); i++ {
		data := c.data[i%len(c.data)]
		if data != "" {
			result = append(result, data)
		}
	}
	return result
}
