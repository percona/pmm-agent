// pmm-agent
// Copyright 2019 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMySQLVersion(t *testing.T) {
	for v, expected := range map[string]string{
		"8.0.24": "8",
		"5.0":    "5",
	} {
		t.Run(v, func(t *testing.T) {
			actual := ParseMySQLVersion(v)
			assert.Equal(t, expected, actual, "%s", v)
		})
	}
}
