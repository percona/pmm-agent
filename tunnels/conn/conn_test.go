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
	"net"
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
	t.Parallel()

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

func TestConn(t *testing.T) {
	t.Parallel()

	tcp1, tcp2, err := pipe()
	require.NoError(t, err)
	defer tcp1.Close()
	defer tcp2.Close()

	var loggerOutput bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&loggerOutput)
	c1 := newConn(tcp1.(*net.TCPConn), logger.WithField("conn", "c1"))
	c2 := newConn(tcp2.(*net.TCPConn), logger.WithField("conn", "c2"))

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

	expected := []byte("hello")
	err = c1.Write(ctx, expected)
	assert.NoError(t, err)
	actual := <-c2.Data()
	assert.Equal(t, expected, actual)

	cancel()
	// wg.Wait()
	assert.Empty(t, loggerOutput.String())
}

// 	// // now check conn implementation
// 	// makePipe = func() (c1, c2 net.Conn, stop func(), err error) {
// 	// 	var tcp1, tcp2 net.Conn
// 	// 	tcp1, tcp2, err = pipe()
// 	// 	if err != nil {
// 	// 		return
// 	// 	}

// 	// 	c1 = newConn(tcp1.(*net.TCPConn))
// 	// 	c2 = newConn(tcp2.(*net.TCPConn))
// 	// 	stop = func() {
// 	// 		_ = c1.Close()
// 	// 		_ = c2.Close()
// 	// 		_ = tcp1.Close()
// 	// 		_ = tcp2.Close()
// 	// 	}
// 	// 	return
// 	// }
// 	// nettest.TestConn(t, makePipe)
// }

// func TestConn(t *testing.T) {
// 	t.Parallel()

// 	c1, src := pipe(t)
// 	dst, c2 := pipe(t)

// 	go func() {
// 		err := copy(dst, src)
// 		require.NoError(t, err)
// 	}()

// 	_, err := fmt.Fprint(c1, "hello\n")
// 	require.NoError(t, err)
// 	var actual string
// 	_, err = fmt.Fscan(c2, &actual)
// 	require.NoError(t, err)
// 	assert.Equal(t, "hello", actual)
// }
