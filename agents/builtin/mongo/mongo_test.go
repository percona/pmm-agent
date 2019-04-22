package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/percona/pmm/api/inventorypb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMongo_Run(t *testing.T) {
	// setup
	l := logrus.WithField("component", "mongo-builtin-agent")
	p := &Params{DSN: "mongodb://127.0.0.1:27017/admin", AgentID: "/agent_id/test"}
	m, err := New(p, l)
	if err != nil {
		t.Fatal(err)
	}

	// run agent
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go m.Run(ctx)

	// collect changes (only check statuses of agent)
	actualStatues := make([]inventorypb.AgentStatus, 0)
	for c := range m.Changes() {
		if c.Status != inventorypb.AgentStatus_AGENT_STATUS_INVALID {
			actualStatues = append(actualStatues, c.Status)
		}
	}

	// waiting agent for sendStopStatus
	<-ctx.Done()

	// check actual statuses with real lifecycle
	expectedStatuses := []inventorypb.AgentStatus{
		inventorypb.AgentStatus_STARTING,
		inventorypb.AgentStatus_RUNNING,
		inventorypb.AgentStatus_STOPPING,
		inventorypb.AgentStatus_DONE,
	}
	assert.Equal(t, expectedStatuses, actualStatues)
}
