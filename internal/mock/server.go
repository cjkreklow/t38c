// Copyright 2020 Collin Kreklow
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
// BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
// ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package mock provides a mock Tile38 server for testing.
package mock

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/tidwall/resp"
)

// Constant values returned by handlers.
const (
	IDNotFound string = "id not found"

	TestOkFalse     string = "test ok false"
	TestServerError string = "test server error"

	TestObject string  = `{"id":"test"}`
	TestTTL    float64 = 5.67
)

// Errors returned by handlers.
var (
	errServerError = errors.New(TestServerError)
)

// Server is a RESP-compatible server for testing.
type Server struct {
	*resp.Server

	Addr string
	Port string

	Err error

	DataIn bytes.Buffer
}

// NewServer returns a new Server listening at Addr:Port.
func NewServer() *Server {
	srv := new(Server)

	srv.Server = resp.NewServer()
	srv.Addr = "127.0.0.1"
	srv.Port = "9876"

	go func(s *Server) {
		err := s.ListenAndServe(net.JoinHostPort(s.Addr, s.Port))
		if err != nil {
			s.Err = err
		}
	}(srv)

	time.Sleep(100 * time.Millisecond) // give server time to start

	return srv
}

// ReturnErr is a handler that returns Err:TestServerError.
func (s *Server) ReturnErr(c *resp.Conn, args []resp.Value) bool {
	var data []byte

	for k, v := range args {
		if k == 0 {
			data = v.Bytes()

			continue
		}

		data = bytes.Join([][]byte{data, v.Bytes()}, []byte(" "))
	}

	s.DataIn.Write(data)

	err := c.WriteError(errServerError)
	if err != nil {
		s.Err = err

		return false
	}

	return true
}

// ReturnOkFalse is a handler that returns Ok:false with Err:TestOkFalse.
func (s *Server) ReturnOkFalse(c *resp.Conn, args []resp.Value) bool {
	var data []byte

	for k, v := range args {
		if k == 0 {
			data = v.Bytes()

			continue
		}

		data = bytes.Join([][]byte{data, v.Bytes()}, []byte(" "))
	}

	s.DataIn.Write(data)

	str := fmt.Sprintf(`{"ok":false,"err":"%s"}`, TestOkFalse)

	err := c.WriteSimpleString(str)
	if err != nil {
		s.Err = err

		return false
	}

	return true
}

// ReturnOkNotFound is a handler that returns Ok:false with
// Err:IdNotFound.
func (s *Server) ReturnOkNotFound(c *resp.Conn, args []resp.Value) bool {
	var data []byte

	for k, v := range args {
		if k == 0 {
			data = v.Bytes()

			continue
		}

		data = bytes.Join([][]byte{data, v.Bytes()}, []byte(" "))
	}

	s.DataIn.Write(data)

	str := fmt.Sprintf(`{"ok":false,"err":"%s"}`, IDNotFound)

	err := c.WriteSimpleString(str)
	if err != nil {
		s.Err = err

		return false
	}

	return true
}

// ReturnOkTrue is a handler that returns Ok:true with
// Object:TestObject and TTL:TestTTL.
func (s *Server) ReturnOkTrue(c *resp.Conn, args []resp.Value) bool {
	var data []byte

	for k, v := range args {
		if k == 0 {
			data = v.Bytes()

			continue
		}

		data = bytes.Join([][]byte{data, v.Bytes()}, []byte(" "))
	}

	s.DataIn.Write(data)

	str := fmt.Sprintf(`{"ok":true,"object":%s,"ttl":%v}`, TestObject, TestTTL)

	err := c.WriteSimpleString(str)
	if err != nil {
		s.Err = err

		return false
	}

	return true
}
