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

const (
	// MaxScanTokenSize is the maximum size used to buffer a token
	// unless the user provides an explicit buffer with Scan.Buffer.
	// The actual maximum token size may be smaller as the buffer
	// may need to include, for instance, a newline.
	MaxScanTokenSize = 128 * 1024

	startBufSize = 4096 // Size of initial allocation for buffer.
)

//CircularWriter is a Writer which write in circular list.
type CircularWriter struct {
	data  []string
	i     int
	l     *logrus.Entry
	m     *sync.RWMutex
	buf   []byte
	start int
	end   int
}

// New creates new CircularWriter.
func New(cap int) *CircularWriter {
	writer := CircularWriter{
		data:  make([]string, cap),
		i:     0,
		l:     logrus.WithField("component", "CircularWriter"),
		m:     &sync.RWMutex{},
		start: 0,
		end:   0,
	}
	return &writer
}

// Write writes new data into buffer.
func (c *CircularWriter) Write(p []byte) (n int, err error) {
	c.m.Lock()
	defer c.m.Unlock()
	// Must read more data.
	// First, shift data to beginning of buffer if there's lots of empty space
	// or space is needed.
	if c.start > 0 && (c.end == len(c.buf) || c.start > len(c.buf)/2) {
		copy(c.buf, c.buf[c.start:c.end])
		c.end -= c.start
		c.start = 0
	}
	// Is the buffer full? If so, resize.
	if c.end == len(c.buf) || c.end+len(p) > len(c.buf) {
		// Guarantee no overflow in the multiplication below.
		const maxInt = int(^uint(0) >> 1)
		if len(c.buf) > maxInt/2 {
			return 0, bufio.ErrTooLong
		}
		newSize := len(c.buf) * 2
		if newSize == 0 {
			newSize = startBufSize
		}
		if newSize > MaxScanTokenSize {
			newSize = MaxScanTokenSize
		}
		newBuf := make([]byte, newSize)
		copy(newBuf, c.buf[c.start:c.end])
		c.buf = newBuf
		c.end -= c.start
		c.start = 0
	}

	n = copy(c.buf[c.end:len(c.buf)], p)
	c.end += n

	for {
		advance, token, err := bufio.ScanLines(c.buf[c.start:c.end], false)
		if err != nil {
			return 0, err
		}
		if token == nil {
			break
		}
		c.i = (c.i + 1) % len(c.data)
		c.data[c.i] = string(token)
		c.start += advance
	}
	return n, nil
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
