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
	"sync"
	"sync/atomic"

	"github.com/percona/pmm/api/agent"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Conn contains business logic of communication with pmm-managed.
type Conn struct {
	stream      agent.Agent_ConnectClient
	lastID      uint32
	l           *logrus.Entry
	rw          sync.RWMutex
	subscribers map[uint32]chan *agent.ServerMessage
	requestChan chan *agent.ServerMessage
}

// NewConn starts goroutine to dispatch messages from server and returns new Conn object
func NewConn(serverAddress string, stream agent.Agent_ConnectClient) *Conn {
	conn := &Conn{
		stream:      stream,
		l:           logrus.WithField("server-address", serverAddress),
		subscribers: make(map[uint32]chan *agent.ServerMessage),
		requestChan: make(chan *agent.ServerMessage),
	}
	// create goroutine to dispatch messages
	go conn.startMessageDispatcher()
	return conn
}

// SendAndRecv sends requests to server and waits for response
func (c *Conn) SendAndRecv(toServer agent.AgentMessagePayload) (*agent.ServerMessage, error) {
	id := atomic.AddUint32(&c.lastID, 1)
	agentMessage := &agent.AgentMessage{
		Id:      id,
		Payload: toServer,
	}
	c.l.Debugf("Send: %s.", agentMessage)
	if err := c.stream.Send(agentMessage); err != nil {
		return nil, errors.Wrap(err, "failed to send message to agent")
	}

	agentChan := make(chan *agent.ServerMessage, 1)

	c.addSubscriber(id, agentChan)

	serverMessage := <-agentChan
	c.l.Debugf("Recv: %s.", serverMessage)

	c.removeSubscriber(id, agentChan)
	close(agentChan)

	return serverMessage, nil
}

// RecvRequestMessage waits for request from server and returns it
func (c *Conn) RecvRequestMessage() *agent.ServerMessage {
	serverMessage := <-c.requestChan
	c.l.Debugf("Recv: %s.", serverMessage)
	return serverMessage
}

func (c *Conn) startMessageDispatcher() {
	context := c.stream.Context()
	for context.Err() == nil {
		serverMessage, err := c.stream.Recv()
		if err != nil {
			c.l.Warnln("Connection is closed", err)
			return
		}
		switch serverMessage.GetPayload().(type) {
		case *agent.ServerMessage_Auth, *agent.ServerMessage_QanData:
			c.emit(serverMessage)
		case *agent.ServerMessage_Ping, *agent.ServerMessage_State:
			go func(serverMessage *agent.ServerMessage) {
				c.requestChan <- serverMessage
			}(serverMessage)
		default:
			c.l.Warnf("unexpected message type %T ", serverMessage.GetPayload())
		}
	}
}

func (c *Conn) emit(message *agent.ServerMessage) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if _, ok := c.subscribers[message.Id]; ok {
		c.subscribers[message.Id] <- message
	} else {
		c.l.Warnf("Unexpected message: %T %s", message, message)
	}
}

func (c *Conn) removeSubscriber(id uint32, subscriber chan *agent.ServerMessage) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if _, ok := c.subscribers[id]; ok {
		delete(c.subscribers, id)
	} else {
		c.l.Warnf("Trying to delete subscriber which is already deleted")
	}
}

func (c *Conn) addSubscriber(id uint32, subscriber chan *agent.ServerMessage) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if _, ok := c.subscribers[id]; ok {
		c.l.Fatalf("Trying to add subscriber to ID which already have subscriber")
	}
	c.subscribers[id] = subscriber
}
