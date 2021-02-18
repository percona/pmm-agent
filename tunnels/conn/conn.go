// pmm-agent
// Copyright 2019 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package conn provides an asynchronous adapter for TCP connection.
package conn

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	readCap            = 42
	writeCap           = 77
	readBuffer         = 4096
	setNoDelay         = true
	setKeepAlivePeriod = 10 * time.Second
	setLinger          = time.Second
	setReadBuffer      = 8192
	setWriteBuffer     = 8192
)

type Conn struct {
	tcp   *net.TCPConn
	l     *logrus.Entry
	read  chan []byte
	write chan []byte
}

func NewConn(tcp *net.TCPConn, l *logrus.Entry) *Conn {
	l = l.WithField("conn", fmt.Sprintf("%s->%s", tcp.LocalAddr(), tcp.RemoteAddr()))
	// set socket options, log warnings on errors
	for _, f := range []func() error{
		func() error { return tcp.SetNoDelay(setNoDelay) },
		func() error { return tcp.SetLinger(int(setLinger.Seconds())) },
		func() error { return tcp.SetReadBuffer(setReadBuffer) },
		func() error { return tcp.SetWriteBuffer(setWriteBuffer) },
		func() error {
			if setKeepAlivePeriod <= 0 {
				return tcp.SetKeepAlive(false)
			}
			if err := tcp.SetKeepAlive(true); err != nil {
				return err //nolint:wrapcheck
			}
			return tcp.SetKeepAlivePeriod(setKeepAlivePeriod)
		},
	} {
		if err := f(); err != nil {
			l.Warn(err)
		}
	}

	return &Conn{
		tcp:   tcp,
		l:     l,
		read:  make(chan []byte, readCap),
		write: make(chan []byte, writeCap),
	}
}

func (c *Conn) Data() <-chan []byte {
	return c.read
}

func (c *Conn) Write(ctx context.Context, b []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.write <- b:
		return nil
	}
}

// Run runs reader and writer until ctx is done or underlying connection is fully closed.
func (c *Conn) Run(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		c.runReader()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		c.runWriter()
		wg.Done()
	}()

	// FIXME connection can close itself "normally", not waiting for ctx
	<-ctx.Done()
	c.l.Debug("ctx Done")

	c.tcp.Close()
	c.l.Debug("tcp closed")

	close(c.write)
	c.l.Debug("write closed")

	wg.Wait()
	c.l.Debug("wait done")

	close(c.read)
	return
}

// runReader reads data from the TCP connection and sends it to the read channel.
// It exits on read error (when connection is closed, for example).
// The caller should close connection and drain the read channel to let runReader return.
func (c *Conn) runReader() {
	// runReader does not accept ctx to have only one way to stop it.

	for {
		b := make([]byte, readBuffer)
		n, err := c.tcp.Read(b)
		if n > 0 {
			c.read <- b[:n]
		}
		if err != nil {
			c.l.Debugf("runReader: %s", err)
			return
		}
	}
}

// runWriter reads data from the write channel and writes it to the TCP connection.
// It exits on write error (when connection is closed, for example) or when the write channel is closed.
func (c *Conn) runWriter() {
	// runWriter does not accept ctx to have only one way to stop it.

	for b := range c.write {
		// TODO Collect several slices, use net.Buffers?

		if _, err := c.tcp.Write(b); err != nil {
			c.l.Debugf("runWriter: %s", err)
			return
		}
	}

	c.l.Debug("runWriter: write channel closed")
}
