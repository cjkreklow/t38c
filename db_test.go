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

package t38c_test

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/tidwall/resp"
	"kreklow.us/go/t38c"
	"kreklow.us/go/t38c/internal/mock"
)

//nolint:gochecknoglobals // internal vars shared between test cases
var (
	srv *mock.Server = mock.NewServer()

	testErrFuncs = map[string][]interface{}{
		"Set":     {"test", "obj1", "STRING", "testing"},
		"Del":     {"test", "obj1"},
		"PDel":    {"test", "obj*"},
		"Expire":  {"test", "obj1", 60},
		"Persist": {"test", "obj1"},
	}

	testRespErrFuncs = map[string][]interface{}{
		"Get":    {"test", "obj1"},
		"Scan":   {"test"},
		"Search": {"test"},
	}
)

// Test Connect errors.
func TestConnectErrors(t *testing.T) {
	t.Run("Fail", testConnectFail)
	t.Run("Error", testConnectError)
	t.Run("False", testConnectFalse)
}

// Test Connect with network error.
func testConnectFail(t *testing.T) {
	db, err := t38c.Connect("fakehost", "9999", 1)
	if err == nil {
		tFatalNoErr(t, "Connect")
	}

	if !strings.HasPrefix(err.Error(), "error connecting to server:") {
		tErrorStr(t, "Connect", "error connecting to servier", err)
	}

	if db != nil {
		tErrorStr(t, "DB", "nil", "not nil")
	}
}

// Test Connect with error from server.
func testConnectError(t *testing.T) {
	srv.HandleFunc("OUTPUT", srv.ReturnErr)
	srv.DataIn.Reset()

	db, err := t38c.Connect("127.0.0.1", "9876", 1)
	if err == nil {
		tFatalNoErr(t, "Connect")
	}

	expErr := fmt.Sprintf("error connecting to server: %s", mock.TestServerError)
	if err.Error() != expErr {
		tErrorStr(t, "Connect", expErr, err)
	}

	if db != nil {
		tErrorStr(t, "DB", "nil", "not nil")
	}

	expData := []byte("OUTPUT json")
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "Data", expData, srv.DataIn.Bytes())
	}
}

// Test Connect with false response from server.
func testConnectFalse(t *testing.T) {
	srv.HandleFunc("OUTPUT", srv.ReturnOkFalse)
	srv.DataIn.Reset()

	db, err := t38c.Connect("127.0.0.1", "9876", 1)
	if err == nil {
		tFatalNoErr(t, "Connect")
	}

	expErr := fmt.Sprintf("error connecting to server: received error: %s", mock.TestOkFalse)
	if err.Error() != expErr {
		tErrorStr(t, "Connect", expErr, err)
	}

	if db != nil {
		tErrorStr(t, "DB", "nil", "not nil")
	}

	expData := []byte("OUTPUT json")
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "Data", expData, srv.DataIn.Bytes())
	}
}

// Test uninitialized errors.
func TestUninitialized(t *testing.T) {
	db := new(t38c.Database)

	expErr := "database not initialized"

	for f, args := range testErrFuncs {
		ret := reflectMethod(db, f, args...)
		if len(ret) != 1 {
			t.Errorf("%s: expected 1 value, received %d", f, len(ret))
		}

		err, ok := ret[0].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T", f, ret[0])
		} else if err.Error() != expErr {
			tErrorStr(t, f, expErr, err)
		}
	}

	for f, args := range testRespErrFuncs {
		ret := reflectMethod(db, f, args...)
		if len(ret) != 2 {
			t.Errorf("%s: expected 2 values, received %d", f, len(ret))
		}

		err, ok := ret[1].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T", f, ret[1])
		} else if err.Error() != expErr {
			tErrorStr(t, f, expErr, err)
		}

		resp, ok := ret[0].(*t38c.Response)
		if !ok {
			t.Errorf("%s: expected *Response, received %T", f, ret[0])
		} else if resp != nil {
			t.Errorf("%s: received non-nil response: %v", f, resp)
		}
	}

	ttl, err := db.TTL("test", "obj1")
	if err == nil {
		tErrorStr(t, "TTL", "error", "nil")
	} else if err.Error() != expErr {
		tErrorStr(t, "TTL", expErr, err)
	}

	if ttl != 0 {
		tErrorVal(t, "TTL", 0, ttl)
	}

	err = db.Close()
	if err == nil {
		tErrorStr(t, "Close", "error", "nil")
	} else if err.Error() != expErr {
		tErrorStr(t, "Close", expErr, err)
	}
}

// Test commands with mock server.
func TestCommands(t *testing.T) {
	t.Run("Server Errors", testCommandErrors)
	t.Run("Response Errors", testCommandFalse)
	t.Run("Set No Args", testCommandSetNoArgs)
	t.Run("Get Not Found", testCommandGetNotFound)
	t.Run("Success", testCommandSuccess)
}

// Test command with error from server.
func testCommandErrors(t *testing.T) {
	testCommandErr(t, srv.ReturnErr, mock.TestServerError)
}

// Test command with false response from server.
func testCommandFalse(t *testing.T) {
	testCommandErr(t, srv.ReturnOkFalse, fmt.Sprintf("received error: %s", mock.TestOkFalse))
}

// Run test commands expecting errors.
func testCommandErr(t *testing.T, hf func(*resp.Conn, []resp.Value) bool, expErr string) { //nolint:funlen,gocognit // long test function okay
	srv.HandleFunc("OUTPUT", srv.ReturnOkTrue)
	srv.DataIn.Reset()

	db, err := t38c.Connect("127.0.0.1", "9876", 1)
	if err != nil {
		tFatalErr(t, "Connect", err)
	}

	if db == nil {
		t.Fatal("Connect: no db returned")
	}

	expData := []byte("OUTPUT json")
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "Data", expData, srv.DataIn.Bytes())
	}

	for f, args := range testErrFuncs {
		cmd := strings.ToUpper(f)
		txtargs := []string{cmd}

		for _, z := range args {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}

		srv.HandleFunc(cmd, hf)
		srv.DataIn.Reset()

		ret := reflectMethod(db, f, args...)
		if len(ret) != 1 {
			t.Errorf("%s: expected 1 value, received %d", f, len(ret))
		}

		err, ok := ret[0].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T", f, ret[0])
		} else if err.Error() != expErr {
			tErrorStr(t, f, expErr, err)
		}

		expData = []byte(strings.Join(txtargs, " "))
		if !bytes.Equal(expData, srv.DataIn.Bytes()) {
			tErrorStr(t, f, expData, srv.DataIn.Bytes())
		}
	}

	for f, args := range testRespErrFuncs {
		cmd := strings.ToUpper(f)
		txtargs := []string{cmd}

		for _, z := range args {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}

		srv.HandleFunc(cmd, hf)
		srv.DataIn.Reset()

		ret := reflectMethod(db, f, args...)
		if len(ret) != 2 {
			t.Errorf("%s: expected 2 values, received %d", f, len(ret))
		}

		err, ok := ret[1].(error)
		if !ok {
			t.Errorf("%s: expected error, received %T", f, ret[1])
		} else if err.Error() != expErr {
			tErrorStr(t, f, expErr, err)
		}

		resp, ok := ret[0].(*t38c.Response)
		if !ok {
			t.Errorf("%s: expected *Response, received %T", f, ret[0])
		} else if resp != nil {
			t.Errorf("%s: received non-nil response: %v", f, resp)
		}

		expData = []byte(strings.Join(txtargs, " "))
		if !bytes.Equal(expData, srv.DataIn.Bytes()) {
			tErrorStr(t, f, expData, srv.DataIn.Bytes())
		}
	}

	srv.HandleFunc("TTL", hf)
	srv.DataIn.Reset()

	ttl, err := db.TTL("test", "obj1")
	if err == nil {
		tErrorStr(t, "TTL", "error", "nil")
	} else if err.Error() != expErr {
		tErrorStr(t, "TTL", expErr, err)
	}

	if ttl != 0 {
		tErrorVal(t, "TTL", 0, ttl)
	}

	expData = []byte("TTL test obj1")
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "TTL", expData, srv.DataIn.Bytes())
	}

	err = db.Close()
	if err != nil {
		tFatalErr(t, "Close", err)
	}
}

// Test Set with no arguments.
func testCommandSetNoArgs(t *testing.T) {
	srv.HandleFunc("OUTPUT", srv.ReturnOkTrue)
	srv.DataIn.Reset()

	db, err := t38c.Connect("127.0.0.1", "9876", 1)
	if err != nil {
		tFatalErr(t, "Connect", err)
	}

	if db == nil {
		t.Fatal("Connect: no db returned")
	}

	expData := []byte("OUTPUT json")
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "Data", expData, srv.DataIn.Bytes())
	}

	srv.HandleFunc("SET", srv.ReturnOkTrue)
	srv.DataIn.Reset()

	expErr := "invalid arguments"

	err = db.Set("test", "obj1")
	if err == nil {
		tErrorStr(t, "Set", "error", "nil")
	} else if err.Error() != expErr {
		tErrorStr(t, "Set", expErr, err)
	}

	expData = []byte{}
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "Set", expData, srv.DataIn.Bytes())
	}

	err = db.Close()
	if err != nil {
		tFatalErr(t, "Close", err)
	}
}

// Test Get not found error.
func testCommandGetNotFound(t *testing.T) {
	srv.HandleFunc("OUTPUT", srv.ReturnOkTrue)
	srv.DataIn.Reset()

	db, err := t38c.Connect("127.0.0.1", "9876", 1)
	if err != nil {
		tFatalErr(t, "Connect", err)
	}

	if db == nil {
		t.Fatal("Connect: no db returned")
	}

	expData := []byte("OUTPUT json")
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "Data", expData, srv.DataIn.Bytes())
	}

	srv.HandleFunc("GET", srv.ReturnOkNotFound)
	srv.DataIn.Reset()

	r, err := db.Get("test", "value1")
	if err != nil {
		tErrorStr(t, "Get", "nil", err)
	}

	if r != nil {
		tErrorVal(t, "Get", "nil", r)
	}

	expData = []byte("GET test value1")
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "Get", expData, srv.DataIn.Bytes())
	}

	err = db.Close()
	if err != nil {
		tFatalErr(t, "Close", err)
	}
}

// Run test commands expecting success.
func testCommandSuccess(t *testing.T) { //nolint:funlen,gocognit // long test function okay
	srv.HandleFunc("OUTPUT", srv.ReturnOkTrue)
	srv.DataIn.Reset()

	db, err := t38c.Connect("127.0.0.1", "9876", 1)
	if err != nil {
		tFatalErr(t, "Connect", err)
	}

	if db == nil {
		t.Fatal("Connect: no db returned")
	}

	expData := []byte("OUTPUT json")
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "Data", expData, srv.DataIn.Bytes())
	}

	expErr := "nope"

	for f, args := range testErrFuncs {
		cmd := strings.ToUpper(f)
		txtargs := []string{cmd}

		for _, z := range args {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}

		srv.HandleFunc(cmd, srv.ReturnOkTrue)
		srv.DataIn.Reset()

		ret := reflectMethod(db, f, args...)
		if len(ret) != 1 {
			t.Errorf("%s: expected 1 value, received %d", f, len(ret))
		}

		err, ok := ret[0].(error)
		if !ok && err != nil {
			t.Errorf("%s: expected error, received %T", f, ret[0])
		} else if err != nil {
			tFatalErr(t, f, err)
		}

		expData = []byte(strings.Join(txtargs, " "))
		if !bytes.Equal(expData, srv.DataIn.Bytes()) {
			tErrorStr(t, f, expData, srv.DataIn.Bytes())
		}
	}

	for f, args := range testRespErrFuncs {
		cmd := strings.ToUpper(f)
		txtargs := []string{cmd}

		for _, z := range args {
			txtargs = append(txtargs, fmt.Sprintf("%v", z))
		}

		srv.HandleFunc(cmd, srv.ReturnOkTrue)
		srv.DataIn.Reset()

		ret := reflectMethod(db, f, args...)
		if len(ret) != 2 {
			t.Errorf("%s: expected 2 values, received %d", f, len(ret))
		}

		err, ok := ret[1].(error)
		if !ok && err != nil {
			t.Errorf("%s: expected error, received %T", f, ret[1])
		} else if err != nil {
			tErrorStr(t, f, expErr, err)
		}

		resp, ok := ret[0].(*t38c.Response)
		if !ok {
			t.Errorf("%s: expected *Response, received %T", f, ret[0])
		} else if resp == nil {
			t.Errorf("%s: received nil response", f)
		}

		if resp.Object != mock.TestObject {
			tErrorStr(t, f, mock.TestObject, resp.Object)
		}

		expData = []byte(strings.Join(txtargs, " "))
		if !bytes.Equal(expData, srv.DataIn.Bytes()) {
			tErrorStr(t, f, expData, srv.DataIn.Bytes())
		}
	}

	srv.HandleFunc("TTL", srv.ReturnOkTrue)
	srv.DataIn.Reset()

	ttl, err := db.TTL("test", "obj1")
	if err != nil {
		tFatalErr(t, "TTL", err)
	}

	if ttl != mock.TestTTL {
		tErrorVal(t, "TTL", mock.TestTTL, ttl)
	}

	expData = []byte("TTL test obj1")
	if !bytes.Equal(expData, srv.DataIn.Bytes()) {
		tErrorStr(t, "TTL", expData, srv.DataIn.Bytes())
	}

	err = db.Close()
	if err != nil {
		tFatalErr(t, "Close", err)
	}
}

// reflectMethod runs a method by reflection.
func reflectMethod(db *t38c.Database, m string, args ...interface{}) []interface{} {
	dbVal := reflect.ValueOf(db)

	argVals := make([]reflect.Value, len(args))
	for i, a := range args {
		argVals[i] = reflect.ValueOf(a)
	}

	f := dbVal.MethodByName(m)
	retVals := f.Call(argVals)

	r := make([]interface{}, len(retVals))
	for i, v := range retVals {
		r[i] = v.Interface()
	}

	return r
}
