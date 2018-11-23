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
	"container/ring"
	"sync"
)

type CircularWriter struct {
	r  *ring.Ring
	rw sync.Mutex
}

func New(len int) *CircularWriter {
	return &CircularWriter{
		r: ring.New(len),
	}
}

func (c *CircularWriter) Write(p []byte) (n int, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.r.Value = p
	c.r = c.r.Next()
	return len(p), nil
}

func (c *CircularWriter) String() string {
	result := ""
	c.r.Do(func(i interface{}) {
		if i != nil {
			result += string(i.([]byte))
		}
	})
	return result
}
