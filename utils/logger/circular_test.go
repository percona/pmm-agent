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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircularWriter(t *testing.T) {
	tests := []struct {
		name        string
		cap         int
		args        []string
		wantData    []string
		expectedLen int
		expectedCap int
	}{
		{
			"simple one",
			4,
			[]string{
				"text\n",
			},
			[]string{"text"},
			0,
			0,
		},
		{
			"two line in one write",
			4,
			[]string{
				"text\nsecond line\n",
			},
			[]string{"text", "second line"},
			0,
			0,
		},
		{
			"three line in two writes",
			4,
			[]string{
				"text\nsecond ",
				"line\nthird row\n",
			},
			[]string{"text", "second line", "third row"},
			0,
			0,
		},
		{
			"log overflow",
			2,
			[]string{
				"text\nsecond ",
				"line\nthird row\n",
			},
			[]string{"second line", "third row"},
			0,
			0,
		},
		{
			"another log overflow",
			2,
			[]string{
				"text\nsecond ",
				"line\nthird row\n",
				"fourth ",
				"line\nlast row\n",
			},
			[]string{"fourth line", "last row"},
			0,
			0,
		},
		{
			"don't write not finished line",
			10,
			[]string{
				"text\nsecond line",
			},
			[]string{"text"},
			11,
			16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.cap)
			for _, arg := range tt.args {
				_, err := c.Write([]byte(arg))
				require.NoError(t, err)
			}
			data := c.Data()
			assert.Equal(t, tt.wantData, data)
			assert.Len(t, c.buf, tt.expectedLen)
			assert.Equal(t, tt.expectedCap, cap(c.buf), "%s", c.buf)
		})
	}
}
