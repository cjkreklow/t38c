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

package t38c

import (
	"errors"
	"fmt"
	"time"

	"github.com/tidwall/gjson"
)

// Define internal errors.
var (
	errInvalidJSON = errors.New("received invalid JSON")
	errDecode      = errors.New("error decoding response")
)

// Response represents a database response.
type Response struct {
	ID          string
	Object      string
	IDs         []string
	Objects     []string
	FieldNames  map[string]int64
	FieldValues []float64
	Count       int64
	Cursor      int64
	TTL         float64
	Err         string
	Elapsed     string
	Ok          bool

	Live    bool
	Command string
	Group   string
	Detect  string
	Key     string
	Time    time.Time

	fields int64
}

// UnmarshalText implements the ability to unmarshal a database
// response.
func (r *Response) UnmarshalText(b []byte) (err error) {
	defer func() {
		// since gjson.ForEach doesn't return errors, catch panics and
		// return them as errors.
		var r = recover()
		if r != nil {
			err = fmt.Errorf("%w: %s", errDecode, r.(string))
		}
	}()

	if !gjson.ValidBytes(b) {
		return errInvalidJSON
	}

	gjson.ParseBytes(b).ForEach(r.decode)

	return nil
}

func (r *Response) decode(k, v gjson.Result) bool {
	switch k.Str {
	case "ok":
		r.Ok = v.Bool()
	case "id":
		r.ID = v.Str
	case "object":
		if v.Type == gjson.JSON {
			r.Object = v.Raw
		} else {
			r.Object = v.Str
		}
	case "ids":
		v.ForEach(func(_, x gjson.Result) bool {
			r.IDs = append(r.IDs, x.Str)

			return true
		})
	case "objects":
		v.ForEach(func(_, x gjson.Result) bool {
			r.Objects = append(r.Objects, x.Raw)

			return true
		})
	case "fields":
		r.decodefields(v)
	case "count":
		r.Count = int64(v.Num)
	case "cursor":
		r.Cursor = int64(v.Num)
	case "ttl":
		r.TTL = v.Num
	case "err":
		r.Err = v.Str
	case "elapsed":
		r.Elapsed = v.Str
	case "live":
		r.Live = v.Bool()
	case "command":
		r.Command = v.Str
	case "group":
		r.Group = v.Str
	case "detect":
		r.Detect = v.Str
	case "key":
		r.Key = v.Str
	case "time":
		r.Time, _ = time.Parse(time.RFC3339Nano, v.Str)
	default:
		panic("unknown response value")
	}

	return true
}

func (r *Response) decodefields(v gjson.Result) {
	r.FieldNames = make(map[string]int64)

	if v.IsArray() {
		v.ForEach(func(_, x gjson.Result) bool {
			r.FieldNames[x.Str] = r.fields
			r.fields++

			return true
		})

		return
	}

	if v.IsObject() {
		v.ForEach(func(l, x gjson.Result) bool {
			r.FieldNames[l.Str] = r.fields
			r.fields++
			r.FieldValues = append(r.FieldValues, x.Num)

			return true
		})

		return
	}

	panic("unknown field type")
}
