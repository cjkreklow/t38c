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

import "testing"

// TestResponseInvalid simulates several error conditions
func TestResponseInvalid(t *testing.T) {
	var err error
	r := new(Response)
	err = r.UnmarshalText([]byte(`{invalid}`))
	if err == nil || err.Error() != "received invalid JSON" {
		t.Errorf("Received: %s\nExpected: received invalid JSON\n", err)
	}

	r = new(Response)
	err = r.UnmarshalText([]byte(`{"zzz":true}`))
	if err == nil || err.Error() != "unknown response value" {
		t.Errorf("Received: %s\nExpected: unknown response value\n", err)
	}

	r = new(Response)
	err = r.UnmarshalText([]byte(`{"fields":true}`))
	if err == nil || err.Error() != "unknown field type" {
		t.Errorf("Received: %s\nExpected: unknown field type\n", err)
	}
}

// TestResponseSingle simulates a single object GET
func TestResponseSingle(t *testing.T) {
	var err error
	r := new(Response)
	err = r.UnmarshalText([]byte(`{"ok":true,"object":{"type":"Point","coordinates":[0,0]},"fields":{"fY":999.999,"fZ":123},"elapsed":"100ms"}`))
	if err != nil {
		t.Errorf("Received unxpected error: %s\n", err)
	}
	if !r.Ok {
		t.Error("Ok not true\n")
	}
	obj := `{"type":"Point","coordinates":[0,0]}`
	if r.Object != obj {
		t.Errorf("Received: %s\nExpected: %s\n", r.Object, obj)
	}
	fields := map[string]float64{
		"fY": 999.999,
		"fZ": 123,
	}
	for k, v := range r.FieldNames {
		if r.FieldValues[v] != fields[k] {
			t.Errorf("Received %s: %f\nExpected: %f\n", k, r.FieldValues[v], fields[k])
		}
	}
	if r.Elapsed != "100ms" {
		t.Errorf("Received: %s\nExpected: %s\n", r.Elapsed, "100ms")
	}
}

// TestResponseMultiple simulates a multiple object SCAN
func TestResponseMultiple(t *testing.T) {
	var err error
	r := new(Response)
	err = r.UnmarshalText([]byte(`{"ok":true,"fields":["fZ","fY"],"objects":[{"id":"value1","object":{"type":"Point","coordinates":[0,0]},"fields":[123,999]},{"id":"value2","object":{"type":"Point","coordinates":[45,45]},"fields":[0,456]}],"count":2,"cursor":1,"elapsed":"567.89µs"}`))
	if err != nil {
		t.Errorf("Received unxpected error: %s\n", err)
	}
	if !r.Ok {
		t.Error("Ok not true\n")
	}
	objs := []string{
		`{"id":"value1","object":{"type":"Point","coordinates":[0,0]},"fields":[123,999]}`,
		`{"id":"value2","object":{"type":"Point","coordinates":[45,45]},"fields":[0,456]}`,
	}
	if r.Objects[0] != objs[0] {
		t.Errorf("Received: %s\nExpected: %s\n", r.Objects[0], objs[0])
	}
	if r.Objects[1] != objs[1] {
		t.Errorf("Received: %s\nExpected: %s\n", r.Objects[1], objs[1])
	}
	if r.Count != 2 {
		t.Errorf("Received: %d\nExpected: %d\n", r.Count, 2)
	}
	if r.Cursor != 1 {
		t.Errorf("Received: %d\nExpected: %d\n", r.Cursor, 1)
	}
	if r.Elapsed != "567.89µs" {
		t.Errorf("Received: %s\nExpected: %s\n", r.Elapsed, "567.89µs")
	}

	r = new(Response)
	err = r.UnmarshalText([]byte(`{"ok":true,"ids":["value1","value2"],"count":2,"cursor":0,"elapsed":"1.7s"}`))
	if err != nil {
		t.Errorf("Received unxpected error: %s\n", err)
	}
	if r.IDs[0] != "value1" {
		t.Errorf("Received: %s\nExpected: %s\n", r.IDs[0], "value1")
	}
	if r.IDs[1] != "value2" {
		t.Errorf("Received: %s\nExpected: %s\n", r.IDs[1], "value2")
	}
}

// TestResponseFence simulates receiving messages from a live geofence
func TestResponseFence(t *testing.T) {
	var err error
	msgs := [][]byte{
		[]byte(`{"ok":true,"live":true}`),
		[]byte(`{"command":"set","group":"5b844beb0a1c1f009ac75639","detect":"enter","key":"test","time":"2018-08-27T19:07:23.578553343Z","id":"value3","object":{"type":"Point","coordinates":[30,30,800]},"fields":{}}`),
		[]byte(`{"command":"set","group":"5b844beb0a1c1f009ac75639","detect":"inside","key":"test","time":"2018-08-27T19:07:23.578553343Z","id":"value3","object":{"type":"Point","coordinates":[30,30,800]},"fields":{}}`),
		[]byte(`{"command":"del","id":"value3","time":"2018-08-27T19:07:33.671191005Z"}`),
	}
	for _, m := range msgs {
		var r = new(Response)
		err = r.UnmarshalText(m)
		if err != nil {
			t.Errorf("Received unxpected error: %s\n", err)
		}
		if !r.Ok && r.Command == "" {
			t.Error("Ok not true and Command not populated\n")
		}
	}
}
