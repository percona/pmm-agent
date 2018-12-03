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
)

type CircularWriter struct {
	data   []string
	i      int
	buffer bytes.Buffer
}

func New(len int) *CircularWriter {
	writer := CircularWriter{
		data:   make([]string, len),
		i:      0,
		buffer: bytes.Buffer{},
	}
	go writer.write()
	return &writer
}

func (c *CircularWriter) write() {
	scanner := bufio.NewScanner(&c.buffer)
	for scanner.Scan() {
		c.data[c.i] = scanner.Text()
		c.i = (c.i + 1) % len(c.data)
	}
}

func (c *CircularWriter) Write(p []byte) (n int, err error) {
	return c.buffer.Write(p)
}

func (c *CircularWriter) Data() []string {
	var result []string
	for i := c.i + 1; i < c.i+len(c.data); i++ {
		result = append(result, c.data[i%len(c.data)])
	}
	return result
}
