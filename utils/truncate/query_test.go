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

package truncate

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/percona/pmm/api/agentpb"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	m := maxQueryLength
	maxQueryLength = 5
	defer func() {
		maxQueryLength = m
	}()

	for q, expected := range map[string]struct {
		query     string
		truncated bool
	}{
		"абвг":    {"абвг", false},
		"абвгд":   {"абвгд", false},
		"абвгде":  {"а ...", true},
		"абвгдеё": {"а ...", true},

		// Unicode replacement characters
		"\xff\xff\xff\xff\xff":     {"\ufffd\ufffd\ufffd\ufffd\ufffd", false},
		"\xff\xff\xff\xff\xff\xff": {"\ufffd ...", true},
	} {
		query, truncated := Query(q)
		assert.Equal(t, expected.query, query)
		assert.Equal(t, expected.truncated, truncated)
	}
}

func TestProtobuf(t *testing.T) {
	query, _ := Query("SELECT * FROM contacts t0 WHERE t0.person_id = '߿�\xff\\uD83D\xdd'")
	bucket := &agentpb.MetricsBucket{
		Common: &agentpb.MetricsBucket_Common{
			Example: query,
		},
	}

	_, err := proto.Marshal(bucket)
	assert.NoError(t, err)
}
