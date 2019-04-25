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

package collector

import (
	"expvar"
)

type stats struct {
	In                     *expvar.Int    `name:"in"`
	Out                    *expvar.Int    `name:"out"`
	IteratorCreated        *expvar.String `name:"iterator-created"`
	IteratorCounter        *expvar.Int    `name:"iterator-counter"`
	IteratorRestartCounter *expvar.Int    `name:"iterator-restart-counter"`
	IteratorErrLast        *expvar.String `name:"iterator-err-last"`
	IteratorErrCounter     *expvar.Int    `name:"iterator-err-counter"`
	IteratorTimeout        *expvar.Int    `name:"iterator-timeout"`
}
