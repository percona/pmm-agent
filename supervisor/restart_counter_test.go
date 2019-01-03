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
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDelay(t *testing.T) {
	rand := rand.New(rand.NewSource(0))
	rc := &restartCounter{
		count: 1,
		rand:  rand,
	}
	for _, expected := range []time.Duration{
		time.Millisecond,
		time.Millisecond,
		4 * time.Millisecond,
		15 * time.Millisecond,
		21 * time.Millisecond,
		time.Millisecond,
		15 * time.Millisecond,
		187 * time.Millisecond,
		464 * time.Millisecond,
	} {
		assert.Equal(t, expected, rc.Delay(), "count = %d", rc.count)
		rc.Inc()
	}
}
