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

// Package mocks contains mocks for tests.
package mocks

import (
	"context"

	"github.com/percona/pmm/api/agent"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type AgentConnectClient struct {
	mock.Mock
	grpc.ClientStream
}

func (x *AgentConnectClient) Context() context.Context {
	ret := x.Called()

	rf := ret.Get(0).(context.Context)

	return rf
}

func (x *AgentConnectClient) Send(m *agent.AgentMessage) error {
	ret := x.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(*agent.AgentMessage) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (x *AgentConnectClient) Recv() (*agent.ServerMessage, error) {
	ret := x.Called()

	var r0 *agent.ServerMessage
	var r1 error
	if rf, ok := ret.Get(0).(func() (*agent.ServerMessage, error)); ok {
		r0, r1 = rf()
	} else {
		r0 = ret.Get(0).(*agent.ServerMessage)
		r1 = ret.Error(1)
	}

	return r0, r1
}

type Context struct {
	mock.Mock
	context.Context
}

func (c *Context) Err() error {
	ret := c.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
