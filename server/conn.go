// pmm-managed
// Copyright (C) 2017 Percona LLC
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

// Package agents contains business logic of working with pmm-agents.
package server

import (
	"sync"
	"sync/atomic"

	"github.com/percona/pmm/api/agent"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Conn struct {
	stream      agent.Agent_ConnectClient
	lastID      uint32
	l           *logrus.Entry
	rw          sync.RWMutex
	subscribers map[uint32][]chan *agent.ServerMessage
	requestChan chan *agent.ServerMessage
}

func NewConn(serverAddress string, stream agent.Agent_ConnectClient) *Conn {
	conn := &Conn{
		stream:      stream,
		l:           logrus.WithField("server-address", serverAddress),
		subscribers: make(map[uint32][]chan *agent.ServerMessage),
		requestChan: make(chan *agent.ServerMessage),
	}
	// create goroutine to dispatch messages
	go conn.startMessageDispatcher()
	return conn
}

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

	agentChan := make(chan *agent.ServerMessage)
	defer close(agentChan)

	c.addSubscriber(id, agentChan)
	defer c.removeSubscriber(id, agentChan)

	serverMessage := <-agentChan
	c.l.Debugf("Recv: %s.", serverMessage)

	return serverMessage, nil
}

func (c *Conn) RecvRequestMessage() *agent.ServerMessage {
	serverMessage := <-c.requestChan
	c.l.Debugf("Recv: %s.", serverMessage)
	return serverMessage
}

func (c *Conn) startMessageDispatcher() {
	for c.stream.Context().Err() != nil {
		serverMessage, err := c.stream.Recv()
		if err != nil {
			c.l.Warnln("Connection closed", err)
			return
		}
		switch serverMessage.GetPayload().(type) {
		case *agent.ServerMessage_Auth, *agent.ServerMessage_QanData:
			c.emit(serverMessage)
		case *agent.ServerMessage_Ping, *agent.ServerMessage_State:
			go func(serverMessage *agent.ServerMessage) {
				c.requestChan <- serverMessage
			}(serverMessage)
			break
		}
	}
}

func (c *Conn) emit(message *agent.ServerMessage) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	if _, ok := c.subscribers[message.Id]; ok {
		for i := range c.subscribers[message.Id] {
			go func(subscriber chan *agent.ServerMessage) {
				subscriber <- message
			}(c.subscribers[message.Id][i])
		}
	} else {
		c.l.Warnf("Unexpected message: %T %s", message, message)
	}
}

func (c *Conn) removeSubscriber(id uint32, messageChan chan *agent.ServerMessage) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if _, ok := c.subscribers[id]; ok {
		for i := range c.subscribers[id] {
			if c.subscribers[id][i] == messageChan {
				c.subscribers[id] = append(c.subscribers[id][:i], c.subscribers[id][i+1:]...)
				break
			}
		}
	}
}

func (c *Conn) addSubscriber(id uint32, agentChan chan *agent.ServerMessage) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if _, ok := c.subscribers[id]; !ok {
		c.subscribers[id] = []chan *agent.ServerMessage{}
	}
	c.subscribers[id] = append(c.subscribers[id], agentChan)
}
