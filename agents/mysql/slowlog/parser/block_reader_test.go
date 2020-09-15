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

package parser

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockReader(t *testing.T) {
	r, err := NewSimpleFileReader(filepath.FromSlash("./testdata/slow012.log"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, r.Close()) })
	br := NewBlockReader(r)

	expected := [][]string{{
		"# User@Host: msandbox[msandbox] @ localhost []  Id:   168\n",
		"# Query_time: 0.000214  Lock_time: 0.000086 Rows_sent: 2  Rows_examined: 2\n",
		"SET timestamp=1397442852;\n",
		"select * from mysql.user;\n",
	}, {
		"# User@Host: msandbox[msandbox] @ localhost []  Id:   168\n",
		"# Query_time: 0.000016  Lock_time: 0.000000 Rows_sent: 2  Rows_examined: 2\n",
		"SET timestamp=1397442852;\n",
		"# administrator command: Quit;\n",
	}, {
		"# Time: 140413 19:34:13\n",
		"# User@Host: msandbox[msandbox] @ localhost [127.0.0.1]  Id:   169\n",
		"# Query_time: 0.000127  Lock_time: 0.000000 Rows_sent: 1  Rows_examined: 0\n",
		"use dev_pct;\n",
		"SET timestamp=1397442853;\n",
		"SELECT @@max_allowed_packet;\n",
	}}

	for i := 0; i < len(expected); i++ {
		actual, err := br.NextBlock()
		assert.Equal(t, expected[i], actual, "block #%d", i)
		require.NoError(t, err, "block #%d", i)
	}

	actual, err := br.NextBlock()
	require.Equal(t, io.EOF, err)
	assert.Nil(t, actual)
}
