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
	type version struct {
		Version         string
		Vendor          string
		ExpectedVersion string
		ExpectedVendor  string
	}

	for _, v := range []version{
		{
			Version:         "8.0",
			Vendor:          "oracle",
			ExpectedVersion: "8",
			ExpectedVendor:  "oracle",
		},
		{
			Version:         "5.7",
			Vendor:          "percona",
			ExpectedVersion: "5",
			ExpectedVendor:  "percona",
		},
	} {
		t.Run(v.Version, func(t *testing.T) {
			version, vendor := ParseMySQLVersion(v.Version, v.Vendor)
			assert.Equal(t, v.ExpectedVersion, version, "%s", v)
			assert.Equal(t, v.ExpectedVendor, vendor, "%s", v)
		})
	}
}
