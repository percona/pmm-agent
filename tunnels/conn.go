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

// import (
// 	"context"
// 	"net"
// 	"time"
// )

// type conn struct {
// 	tcp *net.TCPConn
// }

// func newConn(tcp *net.TCPConn) *conn {
// 	return &conn{
// 		tcp: tcp,
// 	}
// }

// func (c *conn) Run(ctx context.Context) {
// 	reader := make(chan error)
// 	go func() {
// 		buf := make([]byte, 4096)
// 		n, err := c.tcp.Read(buf)
// 		if err != nil {
// 			reader <- err
// 		}
// 	}()

// 	writer := make(chan error)
// 	go func() {
// 		buf := make([]byte, 4096)
// 		n, err := c.tcp.Read(buf)
// 		if err != nil {
// 			reader <- err
// 		}
// 	}()

// 	return
// }

// // Read reads data from the connection.
// // Read can be made to time out and return an error after a fixed
// // time limit; see SetDeadline and SetReadDeadline.
// func (c *conn) Read(b []byte) (n int, err error) {
// 	panic("not implemented") // TODO: Implement
// }

// // Write writes data to the connection.
// // Write can be made to time out and return an error after a fixed
// // time limit; see SetDeadline and SetWriteDeadline.
// func (c *conn) Write(b []byte) (n int, err error) {
// 	panic("not implemented") // TODO: Implement
// }

// // Close closes the connection.
// // Any blocked Read or Write operations will be unblocked and return errors.
// func (c *conn) Close() error {
// 	panic("not implemented") // TODO: Implement
// }

// // LocalAddr returns the local network address.
// func (c *conn) LocalAddr() net.Addr {
// 	panic("not implemented") // TODO: Implement
// }

// // RemoteAddr returns the remote network address.
// func (c *conn) RemoteAddr() net.Addr {
// 	panic("not implemented") // TODO: Implement
// }

// // SetDeadline sets the read and write deadlines associated
// // with the connection. It is equivalent to calling both
// // SetReadDeadline and SetWriteDeadline.
// //
// // A deadline is an absolute time after which I/O operations
// // fail instead of blocking. The deadline applies to all future
// // and pending I/O, not just the immediately following call to
// // Read or Write. After a deadline has been exceeded, the
// // connection can be refreshed by setting a deadline in the future.
// //
// // If the deadline is exceeded a call to Read or Write or to other
// // I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// // This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// // The error's Timeout method will return true, but note that there
// // are other possible errors for which the Timeout method will
// // return true even if the deadline has not been exceeded.
// //
// // An idle timeout can be implemented by repeatedly extending
// // the deadline after successful Read or Write calls.
// //
// // A zero value for t means I/O operations will not time out.
// func (c *conn) SetDeadline(t time.Time) error {
// 	panic("not implemented") // TODO: Implement
// }

// // SetReadDeadline sets the deadline for future Read calls
// // and any currently-blocked Read call.
// // A zero value for t means Read will not time out.
// func (c *conn) SetReadDeadline(t time.Time) error {
// 	panic("not implemented") // TODO: Implement
// }

// // SetWriteDeadline sets the deadline for future Write calls
// // and any currently-blocked Write call.
// // Even if write times out, it may return n > 0, indicating that
// // some of the data was successfully written.
// // A zero value for t means Write will not time out.
// func (c *conn) SetWriteDeadline(t time.Time) error {
// 	panic("not implemented") // TODO: Implement
// }

// // // copy reads bytes from src and writes them dst.
// // // On error, src's read side and dst's write side are closed, and wrapped error is returned.
// // func copy(dst, src *net.TCPConn) error {
// // 	// TODO context, deadlines

// // 	_, err := io.Copy(dst, src)
// // 	_ = src.CloseRead()
// // 	_ = dst.CloseWrite()
// // 	return errors.Wrap(err, "copy")
// // }
