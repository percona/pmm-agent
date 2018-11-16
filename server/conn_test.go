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

// Package server contains business logic of working with pmm-managed.
package server

import (
	"fmt"
	"testing"

	"github.com/percona/pmm-agent/mocks"
	"github.com/percona/pmm/api/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConn_SendAndRecv(t *testing.T) {
	tests := []struct {
		name           string
		request        agent.AgentMessagePayload
		serverMessages []*agent.ServerMessage
		wantId         uint32
		wantType       agent.ServerMessagePayload
		wantErr        bool
	}{
		{
			name:    "Simple test",
			request: &agent.AgentMessage_Auth{},
			serverMessages: []*agent.ServerMessage{{
				Id:      1,
				Payload: &agent.ServerMessage_Auth{},
			}},
			wantId:   1,
			wantType: &agent.ServerMessage_Auth{},
			wantErr:  false,
		},
		{
			name:    "Receiving other response before our one",
			request: &agent.AgentMessage_QanData{},
			serverMessages: []*agent.ServerMessage{{
				Id:      2,
				Payload: &agent.ServerMessage_Auth{},
			}, {
				Id:      1,
				Payload: &agent.ServerMessage_QanData{},
			}},
			wantId:   1,
			wantType: &agent.ServerMessage_QanData{},
			wantErr:  false,
		},
		{
			name:    "Race condition", // Idk when in real life it may become
			request: &agent.AgentMessage_QanData{},
			serverMessages: []*agent.ServerMessage{{
				Id:      1,
				Payload: &agent.ServerMessage_QanData{},
			}, {
				Id:      1,
				Payload: &agent.ServerMessage_QanData{},
			}},
			wantId:   1,
			wantType: &agent.ServerMessage_QanData{},
			wantErr:  false,
		},
		{
			name:    "Receiving request and response with the same id",
			request: &agent.AgentMessage_Auth{},
			serverMessages: []*agent.ServerMessage{{
				Id:      1,
				Payload: &agent.ServerMessage_Ping{},
			}, {
				Id:      1,
				Payload: &agent.ServerMessage_Auth{},
			}},
			wantId:   1,
			wantType: &agent.ServerMessage_Auth{},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := &mocks.AgentConnectClient{}
			ctx := &mocks.Context{}
			currentMessageId := 0

			stream.On("Send", mock.Anything).Return(nil)
			stream.On("Recv").Return(func() (*agent.ServerMessage, error) {
				if currentMessageId < len(tt.serverMessages) {
					serverMessage := tt.serverMessages[currentMessageId]
					currentMessageId++
					return serverMessage, nil
				} else {
					return nil, fmt.Errorf("connection is closed")
				}
			})
			stream.On("Context").Return(ctx)
			ctx.On("Err").Return(nil)

			c := NewConn("mock-server", stream)
			got, err := c.SendAndRecv(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conn.SendAndRecv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Id != tt.wantId {
				t.Errorf("Conn.SendAndRecv() = %v, want %v", got.Id, tt.wantId)
			}
			assert.IsTypef(t, tt.wantType, got.Payload, "Conn.SendAndRecv() payload type = %v, wantType %v")
			ctx.On("Err").Return(fmt.Errorf("stop"))
		})
	}
}

func TestConn_RecvRequestMessage(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []*agent.ServerMessage
		wantId         uint32
		wantType       agent.ServerMessagePayload
	}{
		{
			name: "Simple test",
			serverMessages: []*agent.ServerMessage{{
				Id:      1,
				Payload: &agent.ServerMessage_Ping{},
			}},
			wantId:   1,
			wantType: &agent.ServerMessage_Ping{},
		},
		{
			name: "Receiving request and response with the same id",
			serverMessages: []*agent.ServerMessage{{
				Id:      1,
				Payload: &agent.ServerMessage_Ping{},
			}, {
				Id:      1,
				Payload: &agent.ServerMessage_Auth{},
			}},
			wantId:   1,
			wantType: &agent.ServerMessage_Ping{},
		},
		{
			name: "Receiving response before request",
			serverMessages: []*agent.ServerMessage{{
				Id:      1,
				Payload: &agent.ServerMessage_Auth{},
			}, {
				Id:      2,
				Payload: &agent.ServerMessage_Ping{},
			}},
			wantId:   2,
			wantType: &agent.ServerMessage_Ping{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			stream := &mocks.AgentConnectClient{}
			ctx := &mocks.Context{}
			currentMessageId := 0

			stream.On("Send", mock.Anything).Return(nil)
			stream.On("Recv").Return(func() (*agent.ServerMessage, error) {
				if currentMessageId < len(tt.serverMessages) {
					serverMessage := tt.serverMessages[currentMessageId]
					currentMessageId++
					return serverMessage, nil
				} else {
					return nil, fmt.Errorf("connection is closed")
				}
			})
			stream.On("Context").Return(ctx)
			ctx.On("Err").Return(nil)

			c := NewConn("mock-server", stream)
			got := c.RecvRequestMessage()
			if got.Id != tt.wantId {
				t.Errorf("Conn.RecvRequestMessage() = %v, want %v", got.Id, tt.wantId)
			}
			assert.IsTypef(t, tt.wantType, got.Payload, "Conn.RecvRequestMessage() payload type = %v, wantType %v")
			ctx.On("Err").Return(fmt.Errorf("stop"))
		})
	}
}
