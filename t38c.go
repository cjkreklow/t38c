// Copyright 2024 Collin Kreklow
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

// Package t38c implements client for working with a Tile38 database.
package t38c

import "fmt"

// t38cError is the error type for the t38c library.
type t38cError struct {
	msg  string // error message string from this library
	werr error  // wrapped error from downstream function
}

// Error returns the string value of an error.
func (e *t38cError) Error() string {
	if e.werr == nil {
		return e.msg
	}

	return e.msg + ": " + e.werr.Error()
}

// Unwrap returns an underlying error if applicable.
func (e *t38cError) Unwrap() error {
	return e.werr
}

// newError returns a new t38cError.
func newError(w error, m string) *t38cError {
	return &t38cError{
		msg:  m,
		werr: w,
	}
}

// newErrorf returns a new t38cError with a Printf-style message.
func newErrorf(w error, m string, v ...interface{}) *t38cError {
	return &t38cError{
		msg:  fmt.Sprintf(m, v...),
		werr: w,
	}
}
