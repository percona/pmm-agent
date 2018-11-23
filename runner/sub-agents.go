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
