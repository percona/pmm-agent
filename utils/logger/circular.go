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
	"bytes"
	"sync"

	"github.com/sirupsen/logrus"
)

//CircularWriter is a Writer which write in circular list.
type CircularWriter struct {
	data   []string
	i      int
	l      *logrus.Entry
	buffer *syncBuffer
}

// New creates new CircularWriter.
func New(len int) *CircularWriter {
	writer := CircularWriter{
		data: make([]string, len),
		i:    0,
		buffer: &syncBuffer{
			b: bytes.Buffer{},
		},
		l: logrus.WithField("component", "CircularWriter"),
	}
	go writer.write()
	return &writer
}

func (c *CircularWriter) write() {
	scanner := bufio.NewScanner(c.buffer)
	for scanner.Scan() {
		c.data[c.i] = scanner.Text()
		c.i = (c.i + 1) % len(c.data)
	}
	if err := scanner.Err(); err != nil {
		c.l.Fatalln("can't read from buffer", err)
	}
}

// Write writes new data into buffer.
func (c *CircularWriter) Write(p []byte) (n int, err error) {
	return c.buffer.Write(p)
}

// Data returns string array from circular list.
func (c *CircularWriter) Data() []string {
	var result []string
	currentIndex := c.i
	for i := currentIndex + 1; i <= currentIndex+len(c.data); i++ {
		result = append(result, c.data[i%len(c.data)])
	}
	return result
}

type syncBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *syncBuffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}
func (b *syncBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}
