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

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArguments(t *testing.T) {

	tests := []struct {
		name     string
		args     map[string]string
		expected []string
	}{
		{
			name: "default",
			args: map[string]string{
				"-test":  "test value",
				"-test2": "test2 value",
				"-test3": "",
			},
			expected: []string{"-test", "test value", "-test2", "test2 value", "-test3"},
		},
		{
			name: "only flags",
			args: map[string]string{
				"test":  "",
				"test2": "",
			},
			expected: []string{"test", "test2"},
		},
		{
			name:     "empty",
			args:     map[string]string{},
			expected: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.expected, parseArguments(tt.args))
		})
	}
}
