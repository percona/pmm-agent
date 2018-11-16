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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/percona/pmm-agent/mocks"
	"github.com/percona/pmm/api/agent"
)

func setup(messages []*agent.ServerMessage) (*mocks.Context, *Conn) {
	stream := &mocks.AgentConnectClient{}
	ctx := &mocks.Context{}
	currentMessageID := 0
	stream.On("Send", mock.Anything).Return(nil)
	stream.On("Recv").Return(func() (*agent.ServerMessage, error) {
		if currentMessageID < len(messages) {
			serverMessage := messages[currentMessageID]
			currentMessageID++
			return serverMessage, nil
		}
		return nil, fmt.Errorf("connection is closed")
	})
	stream.On("Context").Return(ctx)
	ctx.On("Err").Return(nil)
	c := NewConn("mock-server", stream)
	return ctx, c
}

func TestConn_SendAndRecv(t *testing.T) {
	tests := []struct {
		name           string
		request        agent.AgentMessagePayload
		serverMessages []*agent.ServerMessage
		wantID         uint32
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
			wantID:   1,
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
			wantID:   1,
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
			wantID:   1,
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
			wantID:   1,
			wantType: &agent.ServerMessage_Auth{},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, conn := setup(tt.serverMessages)
			got, err := conn.SendAndRecv(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conn.SendAndRecv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Id != tt.wantID {
				t.Errorf("Conn.SendAndRecv() = %v, want %v", got.Id, tt.wantID)
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
		wantID         uint32
		wantType       agent.ServerMessagePayload
	}{
		{
			name: "Simple test",
			serverMessages: []*agent.ServerMessage{{
				Id:      1,
				Payload: &agent.ServerMessage_Ping{},
			}},
			wantID:   1,
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
			wantID:   1,
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
			wantID:   2,
			wantType: &agent.ServerMessage_Ping{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, conn := setup(tt.serverMessages)
			got := conn.RecvRequestMessage()
			if got.Id != tt.wantID {
				t.Errorf("Conn.RecvRequestMessage() = %v, want %v", got.Id, tt.wantID)
			}
			assert.IsTypef(t, tt.wantType, got.Payload, "Conn.RecvRequestMessage() payload type = %v, wantType %v")
			ctx.On("Err").Return(fmt.Errorf("stop"))
		})
	}
}
