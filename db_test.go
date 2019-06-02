// Copyright 2019 Collin Kreklow
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

package t38c

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tidwall/resp"
)

var testT *testing.T

var testErrCmds = map[string][]interface{}{
	"Set":     {"test", "obj1", "STRING", "testing"},
	"Del":     {"test", "obj1"},
	"PDel":    {"test", "obj*"},
	"Expire":  {"test", "obj1", 60},
	"Persist": {"test", "obj1"},
}

var testRespErrCmds = map[string][]interface{}{
	"Get":    {"test", "obj1"},
	"Scan":   {"test"},
	"Search": {"test"},
}

// TestUninitialized tests running functions on an uninitialized
// Database object.
func TestUninitialized(t *testing.T) {
	db := new(Database)
	errTxt := "database not initialized"

	for c, a := range testErrCmds {
		ai := reflectMethod(t, db, c, a...)
		if len(ai) != 1 {
			t.Errorf("%s: unexpected return values: %v\n", c, ai)
		}

		err, ok := ai[0].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T\n", c, ai[0])
		} else if err.Error() != errTxt {
			t.Errorf("%s: expected: %s\nreceived: %s\n", c, errTxt, err)
		}
	}

	for c, a := range testRespErrCmds {
		ai := reflectMethod(t, db, c, a...)
		if len(ai) != 2 {
			t.Errorf("%s: unexpected return values: %v\n", c, ai)
		}

		resp, ok := ai[0].(*Response)
		if !ok {
			t.Errorf("%s: expected *Response, received %T\n", c, ai[0])
		} else if resp != nil {
			t.Errorf("%s: received non-nil response: %v\n", c, resp)
		}

		err, ok := ai[1].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T\n", c, ai[1])
		} else if err.Error() != errTxt {
			t.Errorf("%s: expected: %s\nreceived: %s\n", c, errTxt, err)
		}
	}

	ttl, err := db.TTL("test", "obj1")
	if err == nil || err.Error() != errTxt {
		t.Errorf("TTL received: %s\nexpected: %s\n", err, errTxt)
	}
	if ttl != 0 {
		t.Error("TTL received non-zero response with error")
	}

	err = db.Close()
	if err == nil || err.Error() != errTxt {
		t.Errorf("Close received: %s\nexpected: %s\n", err, errTxt)
	}
}

var testport string
var testsrvdata [][]byte
var testsrv *resp.Server
var testdb *Database

// TestDBFunc tests database functions against a mock server.
func TestDBFunc(t *testing.T) {
	// obtain random port
	rand.Seed(time.Now().UnixNano())
	testport = strconv.Itoa(rand.Intn(40000) + 10000)

	// test connection failure
	t.Run("ConnectFail", testConnectFail)

	// start mock server on random port
	testsrv = resp.NewServer()
	go func() {
		err := testsrv.ListenAndServe(net.JoinHostPort("localhost", testport))
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // give server time to start

	// test connection errors
	t.Run("ConnectError", testConnectError)
	t.Run("ConnectFalse", testConnectFalse)

	// test successful connection
	testsrv.HandleFunc("OUTPUT", respReturnOKTrue)
	testsrvdata = [][]byte{}
	output := []byte("OUTPUT json")
	db, err := Connect("127.0.0.1", testport, 1)
	if db == nil {
		t.Fatal("db returned nil")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if len(testsrvdata) != 1 {
		t.Fatalf("received %d values, expected 1\n", len(testsrvdata))
	}
	if !bytes.Equal(testsrvdata[0], []byte("OUTPUT json")) {
		t.Fatalf("Received %s\nExpected: %s\n", testsrvdata, output)
	}

	testdb = db

	// test error responses
	t.Run("CommandError", testCmdError)
	t.Run("CommandFalse", testCmdFalse)
	t.Run("CommandSuccess", testCmdSuccess)
	t.Run("GetNotFound", testGetNotFound)

	err = db.Set("test", "obj")
	if err == nil || err.Error() != "invalid arguments" {
		t.Errorf("Set expected invalid arguments, received %v\n", err)
	}

	r, err := db.runcmd("test")
	if err == nil || err.Error() != "invalid arguments" {
		t.Errorf("runcmd expected invalid arguments, received %v\n", err)
	}
	if r != nil {
		t.Errorf("runcmd expected nil, received response %v\n", r)
	}

	db.Close()
}

// test Connect with no server running
func testConnectFail(t *testing.T) {
	db, err := Connect("127.0.0.1", testport, 1)
	if db != nil {
		t.Error("db not nil on error")
	}
	if err == nil || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("unexpected error received: %v\n", err)
	}
}

// test Connect with error from server
func testConnectError(t *testing.T) {
	output := []byte("OUTPUT json")
	testsrv.HandleFunc("OUTPUT", respReturnErr)
	testsrvdata = [][]byte{}
	db, err := Connect("127.0.0.1", testport, 1)
	if db != nil {
		t.Error("db not nil on error")
	}
	errTxt := "test server error"
	if err == nil || err.Error() != errTxt {
		t.Errorf("Received: %v\nExpected: %s\n", err, errTxt)
	}
	if len(testsrvdata) != 1 {
		t.Errorf("Received %d values, expected 1\n", len(testsrvdata))
	} else if !bytes.Equal(testsrvdata[0], output) {
		t.Errorf("Received %s\nExpected: %s\n", testsrvdata[0], output)
	}
}

// test Connect with false response from server
func testConnectFalse(t *testing.T) {
	output := []byte("OUTPUT json")
	testsrv.HandleFunc("OUTPUT", respReturnOKFalse)
	testsrvdata = [][]byte{}
	db, err := Connect("127.0.0.1", testport, 1)
	if db != nil {
		t.Error("db not nil on error")
	}
	errTxt := "test ok error"
	if err == nil || err.Error() != errTxt {
		t.Errorf("Received: %v\nExpected: %s\n", err, errTxt)
	}
	if len(testsrvdata) != 1 {
		t.Errorf("Received %d values, expected 1\n", len(testsrvdata))
	} else if !bytes.Equal(testsrvdata[0], output) {
		t.Errorf("Received %s\nExpected: %s\n", testsrvdata[0], output)
	}
}

// test error responses
func testCmdError(t *testing.T) {
	errTxt := "test server error"
	for c, a := range testErrCmds {
		testsrvdata = [][]byte{}
		cmd := strings.ToUpper(c)
		txtargs := []string{cmd}
		for _, z := range a {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}
		output := []byte(strings.Join(txtargs, " "))
		testsrv.HandleFunc(cmd, respReturnErr)
		ai := reflectMethod(t, testdb, c, a...)
		if len(ai) != 1 {
			t.Errorf("%s: unexpected return values: %v\n", c, ai)
		}
		err, ok := ai[0].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T\n", c, ai[0])
		} else if err.Error() != errTxt {
			t.Errorf("%s: expected: %s\nreceived: %s\n", c, errTxt, err)
		}
		if len(testsrvdata) != 1 {
			t.Errorf("%s: received %d values, expected 1\n", c, len(testsrvdata))
		} else if !bytes.Equal(testsrvdata[0], output) {
			t.Errorf("%s: received %s\nExpected: %s\n", c, testsrvdata, output)
		}
	}

	for c, a := range testRespErrCmds {
		testsrvdata = [][]byte{}
		cmd := strings.ToUpper(c)
		txtargs := []string{cmd}
		for _, z := range a {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}
		output := []byte(strings.Join(txtargs, " "))
		testsrv.HandleFunc(cmd, respReturnErr)
		ai := reflectMethod(t, testdb, c, a...)
		if len(ai) != 2 {
			t.Errorf("%s: unexpected return values: %v\n", c, ai)
		}
		resp, ok := ai[0].(*Response)
		if !ok {
			t.Errorf("%s: expected *Response, received %T\n", c, ai[0])
		} else if resp != nil {
			t.Errorf("%s: received non-nil response: %v\n", c, resp)
		}
		err, ok := ai[1].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T\n", c, ai[1])
		} else if err.Error() != errTxt {
			t.Errorf("%s: expected: %s\nreceived: %s\n", c, errTxt, err)
		}
		if len(testsrvdata) != 1 {
			t.Errorf("%s: received %d values, expected 1\n", c, len(testsrvdata))
		} else if !bytes.Equal(testsrvdata[0], output) {
			t.Errorf("%s: received %s\nExpected: %s\n", c, testsrvdata, output)
		}
	}

	testsrvdata = [][]byte{}
	output := []byte("TTL test obj1")
	testsrv.HandleFunc("TTL", respReturnErr)
	ttl, err := testdb.TTL("test", "obj1")
	if err == nil || err.Error() != errTxt {
		t.Errorf("TTL received: %s\nexpected: %s\n", err, errTxt)
	}
	if ttl != 0 {
		t.Error("TTL received non-zero response with error")
	}
	if len(testsrvdata) != 1 {
		t.Errorf("TTL received %d values, expected 1\n", len(testsrvdata))
	} else if !bytes.Equal(testsrvdata[0], output) {
		t.Errorf("TTL received %s\nExpected: %s\n", testsrvdata, output)
	}
}

// test false responses
func testCmdFalse(t *testing.T) {
	errTxt := "test ok error"
	for c, a := range testErrCmds {
		testsrvdata = [][]byte{}
		cmd := strings.ToUpper(c)
		txtargs := []string{cmd}
		for _, z := range a {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}
		output := []byte(strings.Join(txtargs, " "))
		testsrv.HandleFunc(cmd, respReturnOKFalse)
		ai := reflectMethod(t, testdb, c, a...)
		if len(ai) != 1 {
			t.Errorf("%s: unexpected return values: %v\n", c, ai)
		}
		err, ok := ai[0].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T\n", c, ai[0])
		} else if err.Error() != errTxt {
			t.Errorf("%s: expected: %s\nreceived: %s\n", c, errTxt, err)
		}
		if len(testsrvdata) != 1 {
			t.Errorf("%s: received %d values, expected 1\n", c, len(testsrvdata))
		} else if !bytes.Equal(testsrvdata[0], output) {
			t.Errorf("%s: received %s\nExpected: %s\n", c, testsrvdata, output)
		}
	}

	for c, a := range testRespErrCmds {
		testsrvdata = [][]byte{}
		cmd := strings.ToUpper(c)
		txtargs := []string{cmd}
		for _, z := range a {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}
		output := []byte(strings.Join(txtargs, " "))
		testsrv.HandleFunc(cmd, respReturnOKFalse)
		ai := reflectMethod(t, testdb, c, a...)
		if len(ai) != 2 {
			t.Errorf("%s: unexpected return values: %v\n", c, ai)
		}
		resp, ok := ai[0].(*Response)
		if !ok {
			t.Errorf("%s: expected *Response, received %T\n", c, ai[0])
		} else if resp != nil {
			t.Errorf("%s: received non-nil response: %v\n", c, resp)
		}
		err, ok := ai[1].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T\n", c, ai[1])
		} else if err.Error() != errTxt {
			t.Errorf("%s: expected: %s\nreceived: %s\n", c, errTxt, err)
		}
		if len(testsrvdata) != 1 {
			t.Errorf("%s: received %d values, expected 1\n", c, len(testsrvdata))
		} else if !bytes.Equal(testsrvdata[0], output) {
			t.Errorf("%s: received %s\nExpected: %s\n", c, testsrvdata, output)
		}
	}

	testsrvdata = [][]byte{}
	output := []byte("TTL test obj1")
	testsrv.HandleFunc("TTL", respReturnOKFalse)
	ttl, err := testdb.TTL("test", "obj1")
	if err == nil || err.Error() != errTxt {
		t.Errorf("TTL received: %s\nexpected: %s\n", err, errTxt)
	}
	if ttl != 0 {
		t.Error("TTL received non-zero response with error")
	}
	if len(testsrvdata) != 1 {
		t.Errorf("TTL received %d values, expected 1\n", len(testsrvdata))
	} else if !bytes.Equal(testsrvdata[0], output) {
		t.Errorf("TTL received %s\nExpected: %s\n", testsrvdata, output)
	}
}

// test success responses
func testCmdSuccess(t *testing.T) {
	for c, a := range testErrCmds {
		testsrvdata = [][]byte{}
		cmd := strings.ToUpper(c)
		txtargs := []string{cmd}
		for _, z := range a {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}
		output := []byte(strings.Join(txtargs, " "))
		testsrv.HandleFunc(cmd, respReturnOKTrue)
		ai := reflectMethod(t, testdb, c, a...)
		if len(ai) != 1 {
			t.Errorf("%s: unexpected return values: %v\n", c, ai)
		}
		err, ok := ai[0].(error)
		if !ok && ai[0] != nil {
			t.Errorf("%s: expected error, received %T\n", c, ai[0])
		} else if err != nil {
			t.Errorf("%s: unexpected error: %v\n", c, err)
		}
		if len(testsrvdata) != 1 {
			t.Errorf("%s: received %d values, expected 1\n", c, len(testsrvdata))
		} else if !bytes.Equal(testsrvdata[0], output) {
			t.Errorf("%s: received %s\nExpected: %s\n", c, testsrvdata, output)
		}
	}

	for c, a := range testRespErrCmds {
		testsrvdata = [][]byte{}
		cmd := strings.ToUpper(c)
		txtargs := []string{cmd}
		for _, z := range a {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}
		output := []byte(strings.Join(txtargs, " "))
		testsrv.HandleFunc(cmd, respReturnOKTrue)
		ai := reflectMethod(t, testdb, c, a...)
		if len(ai) != 2 {
			t.Errorf("%s: unexpected return values: %v\n", c, ai)
		}
		resp, ok := ai[0].(*Response)
		if !ok {
			t.Errorf("%s: expected *Response, received %T\n", c, ai[0])
		} else if resp == nil {
			t.Errorf("%s: received nil response\n", c)
		}
		err, ok := ai[1].(error)
		if !ok && ai[1] != nil {
			t.Errorf("%s: expected error, received %T\n", c, ai[1])
		} else if err != nil {
			t.Errorf("%s: unexpected error: %v\n", c, err)
		}
		if len(testsrvdata) != 1 {
			t.Errorf("%s: received %d values, expected 1\n", c, len(testsrvdata))
		} else if !bytes.Equal(testsrvdata[0], output) {
			t.Errorf("%s: received %s\nExpected: %s\n", c, testsrvdata, output)
		}
	}

	testsrvdata = [][]byte{}
	output := []byte("TTL test obj1")
	testsrv.HandleFunc("TTL", respReturnOKTrue)
	ttl, err := testdb.TTL("test", "obj1")
	if err != nil {
		t.Errorf("TTL unexpected error: %v\n", err)
	}
	if ttl != 5.67 {
		t.Error("TTL received incorrect response")
	}
	if len(testsrvdata) != 1 {
		t.Errorf("TTL received %d values, expected 1\n", len(testsrvdata))
	} else if !bytes.Equal(testsrvdata[0], output) {
		t.Errorf("TTL received %s\nExpected: %s\n", testsrvdata, output)
	}
}

// test Database.Get not found
func testGetNotFound(t *testing.T) {
	output := []byte("GET test value1")
	testsrv.HandleFunc("GET", respReturnOKNotFound)
	testsrvdata = [][]byte{}
	resp, err := testdb.Get("test", "value1")
	if resp != nil {
		t.Errorf("received unexpected response: %v\n", resp)
	}
	if err != nil {
		t.Errorf("received unexpected error: %v\n", err)
	}
	if len(testsrvdata) != 1 {
		t.Errorf("Received %d values, expected 1\n", len(testsrvdata))
	} else if !bytes.Equal(testsrvdata[0], output) {
		t.Errorf("Received %s\nExpected: %s\n", testsrvdata, output)
	}
}

// server handler, returns error
func respReturnErr(c *resp.Conn, args []resp.Value) bool {
	var err error
	var data []byte
	for k, v := range args {
		if k == 0 {
			data = v.Bytes()
			continue
		}
		data = bytes.Join([][]byte{data, v.Bytes()}, []byte(" "))
	}
	testsrvdata = append(testsrvdata, data)
	err = c.WriteError(errors.New("test server error"))
	if err != nil {
		testT.Fatalf("internal write error: %s\n", err)
	}
	return true
}

// server handler, returns ok:false with an err string
func respReturnOKFalse(c *resp.Conn, args []resp.Value) bool {
	var err error
	var data []byte
	for k, v := range args {
		if k == 0 {
			data = v.Bytes()
			continue
		}
		data = bytes.Join([][]byte{data, v.Bytes()}, []byte(" "))
	}
	testsrvdata = append(testsrvdata, data)
	err = c.WriteSimpleString(`{"ok":false,"err":"test ok error"}`)
	if err != nil {
		testT.Fatalf("internal write error: %s\n", err)
	}
	return true
}

// server handler, returns ok:false with a not found error
func respReturnOKNotFound(c *resp.Conn, args []resp.Value) bool {
	var err error
	var data []byte
	for k, v := range args {
		if k == 0 {
			data = v.Bytes()
			continue
		}
		data = bytes.Join([][]byte{data, v.Bytes()}, []byte(" "))
	}
	testsrvdata = append(testsrvdata, data)
	err = c.WriteSimpleString(`{"ok":false,"err":"id not found"}`)
	if err != nil {
		testT.Fatalf("internal write error: %s\n", err)
	}
	return true
}

// server handler, returns ok:true
func respReturnOKTrue(c *resp.Conn, args []resp.Value) bool {
	var err error
	var data []byte
	for k, v := range args {
		if k == 0 {
			data = v.Bytes()
			continue
		}
		data = bytes.Join([][]byte{data, v.Bytes()}, []byte(" "))
	}
	testsrvdata = append(testsrvdata, data)
	err = c.WriteSimpleString(`{"ok":true, "object":"test", "ttl": 5.67}`)
	if err != nil {
		testT.Fatalf("internal write error: %s\n", err)
	}
	return true
}

// reflectMethod runs a method by reflection
func reflectMethod(t *testing.T, db *Database, m string, arg ...interface{}) []interface{} {
	dbv := reflect.ValueOf(db)

	va := make([]reflect.Value, len(arg))
	for i, v := range arg {
		va[i] = reflect.ValueOf(v)
	}

	f := dbv.MethodByName(m)
	rv := f.Call(va)

	r := make([]interface{}, len(rv))
	for i, v := range rv {
		r[i] = v.Interface()
	}

	return r
}
