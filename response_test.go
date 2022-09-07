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
	"testing"

	"kreklow.us/go/t38c"
)

// TestResponseErrors tests errors returned from Response.
func TestResponseErrors(t *testing.T) {
	t.Run("Invalid JSON", testResponseErrJSON)
	t.Run("Unknown Value", testResponseErrValue)
	t.Run("Unknown Type", testResponseErrType)
}

func testResponseErrJSON(t *testing.T) {
	json := []byte(`{invalid}`)
	err := "error unmarshaling response: not valid JSON"

	testResponseErr(t, json, err)
}

func testResponseErrValue(t *testing.T) {
	json := []byte(`{"zzz": true}`)
	err := "error unmarshaling response: unknown response value"

	testResponseErr(t, json, err)
}

func testResponseErrType(t *testing.T) {
	json := []byte(`{"fields": true}`)
	err := "error unmarshaling response: unknown field type"

	testResponseErr(t, json, err)
}

func testResponseErr(t *testing.T, json []byte, e string) {
	t.Helper()

	r := new(t38c.Response)

	err := r.UnmarshalText(json)
	if err == nil {
		tErrorStr(t, "UnmarshalText", "error", "nil")
	} else if e != err.Error() {
		tErrorStr(t, "UnmarshalText", e, err)
	}
}

// TestResponseValid tests valid Responses.
func TestResponseValid(t *testing.T) {
	t.Run("Single JSON", testResponseSingleJSON)
	t.Run("Single String", testResponseSingleString)
	t.Run("Multiple Objects", testResponseMultObj)
	t.Run("Multiple IDs", testResponseMultID)
	t.Run("Geofence", testResponseFence)
	t.Run("Server Error", testResponseSrvErr)
}

func testResponseSingleJSON(t *testing.T) {
	json := []byte(`{"ok":true,"object":{"type":"Point","coordinates":[0,0]},"fields":{"fY":999.999,"fZ":123},"ttl":500,"elapsed":"100ms"}`)
	obj := `{"type":"Point","coordinates":[0,0]}`

	testResponseSingle(t, json, obj)
}

func testResponseSingleString(t *testing.T) {
	json := []byte(`{"ok":true,"object":"objstr","fields":{"fY":999.999,"fZ":123},"ttl":500,"elapsed":"100ms"}`)
	obj := `objstr`

	testResponseSingle(t, json, obj)
}

func testResponseSingle(t *testing.T, json []byte, obj string) {
	t.Helper()

	r := new(t38c.Response)

	err := r.UnmarshalText(json)
	if err != nil {
		tFatalErr(t, "UnmarshalText", err)
	}

	if !r.Ok {
		tErrorStr(t, "Ok", "true", "false")
	}

	if r.Object != obj {
		tErrorStr(t, "Object", obj, r.Object)
	}

	fields := map[string]float64{
		"fY": 999.999,
		"fZ": 123,
	}
	for k, v := range r.FieldNames {
		if r.FieldValues[v] != fields[k] {
			tErrorVal(t, k, fields[k], r.FieldValues[v])
		}
	}

	expTTL := float64(500)
	if r.TTL != expTTL {
		tErrorVal(t, "TTL", expTTL, r.TTL)
	}

	expElapsed := "100ms"
	if r.Elapsed != expElapsed {
		tErrorStr(t, "Elapsed", expElapsed, r.Elapsed)
	}
}

func testResponseMultObj(t *testing.T) {
	json := []byte(`{"ok":true,"fields":["fZ","fY"],"objects":[{"id":"value1","object":{"type":"Point","coordinates":[0,0]},"fields":[123,999]},{"id":"value2","object":{"type":"Point","coordinates":[45,45]},"fields":[0,456]}],"count":2,"cursor":0,"elapsed":"567.89µs"}`)
	objs := []string{
		`{"id":"value1","object":{"type":"Point","coordinates":[0,0]},"fields":[123,999]}`,
		`{"id":"value2","object":{"type":"Point","coordinates":[45,45]},"fields":[0,456]}`,
	}

	testResponseMult(t, json, objs, nil)
}

func testResponseMultID(t *testing.T) {
	json := []byte(`{"ok":true,"ids":["value1","value2"],"count":2,"cursor":0,"elapsed":"567.89µs"}`)
	ids := []string{"value1", "value2"}

	testResponseMult(t, json, nil, ids)
}

func testResponseMult(t *testing.T, json []byte, objs []string, ids []string) {
	t.Helper()

	r := new(t38c.Response)

	err := r.UnmarshalText(json)
	if err != nil {
		tFatalErr(t, "UnmarshalText", err)
	}

	if !r.Ok {
		tErrorStr(t, "Ok", "true", "false")
	}

	if objs != nil {
		if len(r.Objects) < 2 {
			t.Fatal("not enough objects returned")
		}

		if r.Objects[0] != objs[0] {
			tErrorStr(t, "Object 0", objs[0], r.Objects[0])
		}

		if r.Objects[1] != objs[1] {
			tErrorStr(t, "Object 1", objs[1], r.Objects[1])
		}
	}

	if ids != nil {
		if len(r.IDs) < 2 {
			t.Fatal("not enough IDs returned")
		}

		if r.IDs[0] != ids[0] {
			tErrorStr(t, "ID 0", ids[0], r.IDs[0])
		}

		if r.IDs[1] != ids[1] {
			tErrorStr(t, "ID 1", ids[1], r.IDs[1])
		}
	}

	expCount := int64(2)
	if r.Count != expCount {
		tErrorVal(t, "Count", expCount, r.Count)
	}

	expCursor := int64(0)
	if r.Cursor != expCursor {
		tErrorVal(t, "Cursor", expCursor, r.Cursor)
	}

	expElapsed := "567.89µs"
	if r.Elapsed != expElapsed {
		tErrorVal(t, "Elapsed", expElapsed, r.Elapsed)
	}
}

func testResponseFence(t *testing.T) {
	msgs := [][]byte{
		[]byte(`{"ok":true,"live":true}`),
		[]byte(`{"command":"set","group":"5b844beb0a1c1f009ac75639","detect":"enter","key":"test","time":"2018-08-27T19:07:23.578553343Z","id":"value3","object":{"type":"Point","coordinates":[30,30,800]},"fields":{}}`),
		[]byte(`{"command":"set","group":"5b844beb0a1c1f009ac75639","detect":"inside","key":"test","time":"2018-08-27T19:07:23.578553343Z","id":"value3","object":{"type":"Point","coordinates":[30,30,800]},"fields":{}}`),
		[]byte(`{"command":"del","id":"value3","time":"2018-08-27T19:07:33.671191005Z"}`),
	}

	for _, m := range msgs {
		r := new(t38c.Response)

		err := r.UnmarshalText(m)
		if err != nil {
			tFatalErr(t, "UnmarshalText", err)
		}

		if !r.Ok && r.Command == "" {
			t.Error("Ok not true and Command not populated")
		}
	}
}

func testResponseSrvErr(t *testing.T) {
	json := []byte(`{"ok":true,"err":"bad request"}`)

	r := new(t38c.Response)

	err := r.UnmarshalText(json)
	if err != nil {
		tFatalErr(t, "UnmarshalText", err)
	}

	if !r.Ok {
		tErrorStr(t, "Ok", "true", "false")
	}

	expErr := "bad request"
	if expErr != r.Err {
		tErrorStr(t, "Err", expErr, r.Err)
	}
}
