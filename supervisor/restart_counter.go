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
	"math"
	"math/rand"
	"sync/atomic"
	"time"
)

type restartCounter struct {
	count int32
}

func (r *restartCounter) Inc() {
	atomic.AddInt32(&r.count, 1)
}

func (r *restartCounter) Reset() {
	atomic.CompareAndSwapInt32(&r.count, r.count, 1)
}

func (r *restartCounter) Delay() time.Duration {
	max := math.Pow(2, float64(r.count)) - 1
	delay := rand.Int63n(int64(max))
	return (1 + time.Duration(delay)) * time.Millisecond
}
