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

package runner

import (
	"context"
	"github.com/percona/pmm/api/inventory"
)

type State int32

const (
	INVALID State = 0
	RUNNING State = 1
	STOPPED State = 2
	CRASHED State = 3
)

type AgentParams struct {
	AgentId uint32
	Type    inventory.AgentType
	Args    []string
	Env     []string
	Configs map[string]string
	Port    uint32
}

type SubAgent interface {
	Start(ctx context.Context) error
	Stop() error
	GetLogs() string
	GetState() State
}
