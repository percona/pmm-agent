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
)

func TestCircularWriter(t *testing.T) {
	tests := []struct {
		name     string
		len      int
		args     [][]byte
		wantData []string
		wantErr  bool
	}{
		{
			"simple one",
			4,
			[][]byte{
				[]byte("text\n"),
			},
			[]string{"text"},
			false,
		},
		{
			"two line in one write",
			4,
			[][]byte{
				[]byte("text\nsecond line\n"),
			},
			[]string{"text", "second line"},
			false,
		},
		{
			"three line in two writes",
			4,
			[][]byte{
				[]byte("text\nsecond "),
				[]byte("line\nthird row\n"),
			},
			[]string{"text", "second line", "third row"},
			false,
		},
		{
			"three line in two writes",
			4,
			[][]byte{
				[]byte("text\nsecond "),
				[]byte("line\nthird row\n"),
			},
			[]string{"text", "second line", "third row"},
			false,
		},
		{
			"log overflow",
			2,
			[][]byte{
				[]byte("text\nsecond "),
				[]byte("line\nthird row\n"),
			},
			[]string{"second line", "third row"},
			false,
		},
		{
			"another log overflow",
			2,
			[][]byte{
				[]byte("text\nsecond "),
				[]byte("line\nthird row\n"),
				[]byte("fourth "),
				[]byte("line\nlast row\n"),
			},
			[]string{"fourth line", "last row"},
			false,
		},
		{
			"don't write not finished line",
			10,
			[][]byte{
				[]byte("text\nsecond line"),
			},
			[]string{"text"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.len)
			for _, arg := range tt.args {
				_, err := c.Write(arg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CircularWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			data := c.Data()
			assert.Equal(t, tt.wantData, data)
		})
	}
}
