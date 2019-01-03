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

package supervisor

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/percona/pmm/api/agent"
)

func TestRaceCondition(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx, cancel := context.WithCancel(context.Background())
	m := newSubAgent(ctx, &agent.SetStateRequest_AgentProcess{
		Type: agent.Type_MYSQLD_EXPORTER,
		Args: []string{"-web.listen-address=127.0.0.1:11111"},
		Env: []string{
			`DATA_SOURCE_NAME="pmm:pmm@(127.0.0.1:3306)/pmm-managed-dev"`,
		},
	})
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	assert.NotPanics(t, func() {
		for {
			state := <-m.Changes()
			if state == STOPPED {
				break
			}
		}
	})
}

func TestStates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := newSubAgent(ctx, &agent.SetStateRequest_AgentProcess{
		Type: agent.Type_MYSQLD_EXPORTER,
		Args: []string{"-web.listen-address=127.0.0.1:11112"},
		Env: []string{
			`DATA_SOURCE_NAME="pmm:pmm@(127.0.0.1:3306)/pmm-managed-dev"`,
		},
	})

	assert.Equal(t, STARTING, <-m.Changes())
	assert.Equal(t, RUNNING, <-m.Changes())
	m.cmdM.Lock()
	err := syscall.Kill(m.cmd.Process.Pid, syscall.SIGKILL)
	m.cmdM.Unlock()
	assert.NoError(t, err)
	assert.Equal(t, WAITING, <-m.Changes())
	assert.Equal(t, STARTING, <-m.Changes())
	assert.Equal(t, RUNNING, <-m.Changes())
	cancel()
	assert.Equal(t, STOPPING, <-m.Changes())
	assert.Equal(t, STOPPED, <-m.Changes())
}

func TestStatesOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := newSubAgent(ctx, &agent.SetStateRequest_AgentProcess{
		Type: agent.Type_MYSQLD_EXPORTER,
		Args: []string{"-web.listen-address=127.0.0.1:11112"},
		Env: []string{
			`DATA_SOURCE_NAME="pmm:pmm@(127.0.0.1:3306)/pmm-managed-dev"`,
		},
	})

	assert.Equal(t, STARTING, <-m.Changes())
	assert.Equal(t, RUNNING, <-m.Changes())
	cancel()
	assert.Equal(t, STOPPING, <-m.Changes())
	assert.Equal(t, STOPPED, <-m.Changes())
}

func TestStopOnStartingState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := newSubAgent(ctx, &agent.SetStateRequest_AgentProcess{
		Type: agent.Type_MYSQLD_EXPORTER,
		Args: []string{"-web.listen-address=127.0.0.1:11112"},
		Env: []string{
			`DATA_SOURCE_NAME="pmm:pmm@(127.0.0.1:3306)/pmm-managed-dev"`,
		},
	})

	assert.Equal(t, STARTING, <-m.Changes())
	cancel()
	time.Sleep(1 * time.Second)
	assert.Equal(t, RUNNING, <-m.Changes())
	assert.Equal(t, STOPPING, <-m.Changes())
	assert.Equal(t, STOPPED, <-m.Changes())
}

func TestNotFoundBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := newSubAgent(ctx, &agent.SetStateRequest_AgentProcess{
		Type: Type_TESTING_NOT_FOUND,
	})

	assert.Equal(t, STARTING, <-m.Changes())
	cancel()
	time.Sleep(1 * time.Second)
}
