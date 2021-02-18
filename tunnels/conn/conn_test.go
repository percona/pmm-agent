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

package conn

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/nettest"
)

// pipe returns two connections that acts like a pipe: reads on one end are matched with writes on the other.
// Unlike net.Pipe(), the returned connections are real TCP connections - they have buffers, etc.
func pipe() (net.Conn, net.Conn, error) {
	l, err0 := net.Listen("tcp4", "127.0.0.1:0")
	if err0 != nil {
		return nil, nil, err0
	}
	defer l.Close()

	var c1 net.Conn
	var err1 error
	accepted := make(chan struct{})
	go func() {
		c1, err1 = l.Accept()
		close(accepted)
	}()

	c2, err2 := net.Dial("tcp4", l.Addr().String())
	if err2 != nil {
		return nil, nil, err2
	}

	<-accepted
	if err1 != nil {
		return nil, nil, err1
	}

	return c1, c2, nil
}

func TestPipe(t *testing.T) {
	makePipe := func() (c1, c2 net.Conn, stop func(), err error) {
		c1, c2, err = pipe()
		stop = func() {
			_ = c1.Close()
			_ = c2.Close()
		}
		return
	}
	nettest.TestConn(t, makePipe)
}

func setup(t *testing.T) (*Conn, *Conn, fmt.Stringer) {
	t.Helper()

	tcp1, tcp2, err := pipe()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tcp1.Close()
		_ = tcp2.Close()
	})

	var loggerOutput bytes.Buffer
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "15:04:05.000000",
	})
	logger.SetOutput(io.MultiWriter(os.Stderr, &loggerOutput))
	logger.SetLevel(logrus.DebugLevel)

	c1 := NewConn(tcp1.(*net.TCPConn), logger.WithField("conn", "c1"))
	c2 := NewConn(tcp2.(*net.TCPConn), logger.WithField("conn", "c2"))
	return c1, c2, &loggerOutput
}

func TestBasic(t *testing.T) {
	c1, c2, log := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		c1.Run(ctx)
		wg.Done()
	}()
	go func() {
		c2.Run(ctx)
		wg.Done()
	}()

	const total = 10 * 1024 * 1024

	// c1 writer
	var wrote int
	go func() {
		// r := rand.New(rand.NewSource(1)) //nolint:gosec
		for {
			b := make([]byte, 1024*1024) // r.Intn(1024*1024))
			wrote += len(b)
			// if wrote > total {
			// 	b = b[:len(b)-(wrote-total)]
			// 	wrote = total
			// }
			err := c1.Write(b)
			if !assert.NoError(t, err) {
				return
			}
			if wrote == total {
				c1.CloseWrite()
				return
			}
		}
	}()

	c2.CloseWrite()

	var read int
	for b := range c2.Data() {
		read += len(b)
	}

	require.Equal(t, wrote, read)

	cancel()
	wg.Wait()

	assert.NotContains(t, log.String(), "level=error")
	assert.NotContains(t, log.String(), "level=warn")
	assert.Contains(t, log.String(), "level=debug")
}
