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
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	readCap            = 42
	writeCap           = 77
	readBuffer         = 64 * 1024
	setNoDelay         = true
	setKeepAlivePeriod = 10 * time.Second
	setLinger          = -1
	setReadBuffer      = 64 * 1024
	setWriteBuffer     = 64 * 1024
)

type Conn struct {
	tcp *net.TCPConn
	l   *logrus.Entry

	read         chan []byte
	write        chan []byte
	ignoreWrites chan struct{}
	writeOnce    sync.Once

	readBytesTotal  uint64
	wroteBytesTotal uint64
}

type Metrics struct {
	ReadBytesTotal             uint64
	WroteBytesTotal            uint64
	ReadChanLen, ReadChanCap   int
	WriteChanLen, WriteChanCap int
}

func NewConn(tcp *net.TCPConn, l *logrus.Entry) *Conn {
	l = l.WithField("local", tcp.LocalAddr().String())
	l = l.WithField("remote", tcp.RemoteAddr().String())

	// set socket options, log warnings on errors
	for _, f := range []func() error{
		func() error { return tcp.SetNoDelay(setNoDelay) },
		func() error { return tcp.SetLinger(setLinger) },
		func() error {
			if setReadBuffer <= 0 {
				return nil
			}
			return tcp.SetReadBuffer(setReadBuffer)
		},
		func() error {
			if setWriteBuffer <= 0 {
				return nil
			}
			return tcp.SetWriteBuffer(setWriteBuffer)
		},
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
		tcp:          tcp,
		l:            l,
		read:         make(chan []byte, readCap),
		write:        make(chan []byte, writeCap),
		ignoreWrites: make(chan struct{}),
	}
}

func (c *Conn) Data() <-chan []byte {
	return c.read
}

func (c *Conn) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	select {
	case <-c.ignoreWrites:
		return fmt.Errorf("ignoreWrites")
	default:
	}

	c.write <- b
	return nil
}

func (c *Conn) CloseWrite() {
	c.writeOnce.Do(func() {
		close(c.ignoreWrites)
	})
}

func (c *Conn) Metrics() *Metrics {
	return &Metrics{
		ReadBytesTotal:  atomic.LoadUint64(&c.readBytesTotal),
		WroteBytesTotal: atomic.LoadUint64(&c.wroteBytesTotal),
		ReadChanLen:     len(c.read),
		ReadChanCap:     cap(c.read),
		WriteChanLen:    len(c.write),
		WriteChanCap:    cap(c.write),
	}
}

// Run runs reader and writer until ctx is done or underlying connection is fully closed.
func (c *Conn) Run(ctx context.Context) {
	var wg sync.WaitGroup

	readerDone := make(chan struct{})
	wg.Add(1)
	go func() {
		c.runReader()
		c.l.Debugf("runReader done, closing read channel")
		close(c.read)
		close(readerDone)
		wg.Done()
	}()

	writerDone := make(chan struct{})
	wg.Add(1)
	go func() {
		c.runWriter()
		c.l.Debugf("runWriter done, closing TCP connection's write side")
		_ = c.tcp.CloseWrite()
		close(writerDone)
		wg.Done()
	}()

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	wg.Add(1)
	go func() {
		<-ctx.Done()
		c.l.Debugf("ctx done, closing TCP connection")
		_ = c.tcp.Close()
		wg.Done()
	}()

	<-writerDone
	c.CloseWrite()
	<-readerDone
	cancel()
	wg.Wait()
}

// runReader reads data from the TCP connection and sends it to the read channel.
// It exits on read error (when connection is closed, for example).
func (c *Conn) runReader() {
	// runReader does not accept ctx to have only one way to stop it.

	for {
		// time.Sleep(25 * time.Millisecond)

		b := make([]byte, readBuffer)
		n, err := c.tcp.Read(b)
		atomic.AddUint64(&c.readBytesTotal, uint64(n))

		log := c.l.Tracef
		if err != nil {
			log = c.l.Debugf
		}
		log("runReader: read %d/%d bytes; channel %d/%d; %v", n, len(b), len(c.read), cap(c.read), err)

		if n > 0 {
			c.read <- b[:n]
		}
		if err != nil {
			return
		}
	}
}

// runWriter reads data from the write channel and writes it to the TCP connection.
// It exits on write error (when connection is closed, for example).
func (c *Conn) runWriter() {
	// runWriter does not accept ctx to have only one way to stop it.

	for {
		select {
		case b := <-c.write:
			if err := c.realWrite(b); err != nil {
				return
			}
			continue
		default:
		}

		select {
		case <-c.ignoreWrites:
			return
		case b := <-c.write:
			if err := c.realWrite(b); err != nil {
				return
			}
		}
	}
}

func (c *Conn) realWrite(b []byte) error {
	n, err := c.tcp.Write(b)
	atomic.AddUint64(&c.wroteBytesTotal, uint64(n))

	log := c.l.Tracef
	if err != nil {
		log = c.l.Debugf
	}
	log("runWriter: wrote %d/%d bytes; channel %d/%d; %v", n, len(b), len(c.write), cap(c.write), err)

	return err
}
