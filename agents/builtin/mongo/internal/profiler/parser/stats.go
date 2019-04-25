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

package parser

import (
	"expvar"
)

type stats struct {
	InDocs         *expvar.Int    `name:"docs-in"`
	OkDocs         *expvar.Int    `name:"docs-ok"`
	OutReports     *expvar.Int    `name:"reports-out"`
	IntervalStart  *expvar.String `name:"interval-start"`
	IntervalEnd    *expvar.String `name:"interval-end"`
	ErrFingerprint *expvar.Int    `name:"err-fingerprint"`
	ErrParse       *expvar.Int    `name:"err-parse"`
	SkippedDocs    *expvar.Int    `name:"skipped-docs"`
}
