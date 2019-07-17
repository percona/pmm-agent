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

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePostgreSQLVersion(t *testing.T) {
	for v, expected := range map[string]string{
		"PostgreSQL 12beta2 (Debian 12~beta2-1.pgdg100+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 8.3.0-6) 8.3.0, 64-bit":              "12",
		"PostgreSQL 10.9 (Debian 10.9-1.pgdg90+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit":     "10",
		"PostgreSQL 9.4.23 on x86_64-pc-linux-gnu (Debian 9.4.23-1.pgdg90+1), compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit": "9.4",
	} {
		t.Run(v, func(t *testing.T) {
			actual := parsePostgreSQLVersion(v)
			assert.Equal(t, expected, actual, "%s", v)
		})
	}
}
