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

package tunnels

// "golang.org/x/net/nettest"

// // pipe returns two connections that acts like a pipe: reads on one end are matched with writes on the other.
// // Unlike net.Pipe(), the returned connections are real TCP connections - they have buffers, etc.
// func pipe() (net.Conn, net.Conn, error) {
// 	l, err0 := net.Listen("tcp4", "127.0.0.1:0")
// 	if err0 != nil {
// 		return nil, nil, err0
// 	}
// 	defer l.Close()

// 	var c1 net.Conn
// 	var err1 error
// 	done := make(chan struct{})
// 	go func() {
// 		c1, err1 = l.Accept()
// 		close(done)
// 	}()

// 	c2, err2 := net.Dial("tcp4", l.Addr().String())
// 	if err2 != nil {
// 		return nil, nil, err2
// 	}

// 	<-done
// 	if err1 != nil {
// 		return nil, nil, err1
// 	}

// 	return c1, c2, nil
// }

// func TestConn(t *testing.T) {
// 	t.Parallel()

// 	// first check that pipe() is not broken
// 	makePipe := func() (c1, c2 net.Conn, stop func(), err error) {
// 		c1, c2, err = pipe()
// 		stop = func() {
// 			_ = c1.Close()
// 			_ = c2.Close()
// 		}
// 		return
// 	}
// 	nettest.TestConn(t, makePipe)

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
