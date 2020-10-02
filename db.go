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
	"net"
	"strconv"

	"github.com/mediocregopher/radix/v3"
)

var (
	errUninitialized = errors.New("database not initialized")
	errArgs          = errors.New("invalid arguments")
	errResponse      = errors.New("received error")
)

// Database is the primary object for interacting with the database.
// Database should not be created directly, instead use Connect() to
// retrieve a fully-initialized Database ready to be used.
//
// Functions other than Close() accept arguments in the same form as the
// Tile38 CLI. See https://tile38.com/commands/ for further information.
//
type Database struct {
	pool *radix.Pool
}

// Connect establishes a connection and returns a Database object.
func Connect(server string, port string, poolsize int) (db *Database, err error) {
	db = new(Database)

	db.pool, err = radix.NewPool(
		"tcp",
		net.JoinHostPort(server, port),
		poolsize,
		radix.PoolConnFunc(connectJSON),
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes the database connection.
func (db *Database) Close() error {
	if db.pool == nil {
		return errUninitialized
	}

	return db.pool.Close()
}

// Set saves an object to the database.
func (db *Database) Set(key string, id string, args ...string) (err error) {
	if db.pool == nil {
		return errUninitialized
	}

	if args == nil {
		return errArgs
	}

	cmdargs := append([]string{key, id}, args...)

	_, err = db.runcmd("SET", cmdargs...)
	if err != nil {
		return err
	}

	return nil
}

// Get returns the requested entry as a response object, or nil if the
// object is not found.
func (db *Database) Get(key string, id string, args ...string) (r *Response, err error) {
	if db.pool == nil {
		return nil, errUninitialized
	}

	cmdargs := []string{key, id}

	if args != nil {
		cmdargs = append(cmdargs, args...)
	}

	r, err = db.runcmd("GET", cmdargs...)
	if err != nil {
		if err.Error() == "id not found" {
			return nil, nil
		}

		return nil, err
	}

	return r, nil
}

// Scan iterates through a key returning a set of results.
func (db *Database) Scan(key string, args ...string) (r *Response, err error) {
	if db.pool == nil {
		return nil, errUninitialized
	}

	cmdargs := []string{key}

	if args != nil {
		cmdargs = append(cmdargs, args...)
	}

	r, err = db.runcmd("SCAN", cmdargs...)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Search iterates through the string values of a key returning a set of
// results.
func (db *Database) Search(key string, args ...string) (r *Response, err error) {
	if db.pool == nil {
		return nil, errUninitialized
	}

	cmdargs := []string{key}

	if args != nil {
		cmdargs = append(cmdargs, args...)
	}

	r, err = db.runcmd("SEARCH", cmdargs...)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Del deletes the requested entry.
func (db *Database) Del(key string, id string) (err error) {
	if db.pool == nil {
		return errUninitialized
	}

	_, err = db.runcmd("DEL", key, id)
	if err != nil {
		return err
	}

	return nil
}

// PDel deletes any entries matching the supplied pattern.
func (db *Database) PDel(key string, pattern string) (err error) {
	if db.pool == nil {
		return errUninitialized
	}

	_, err = db.runcmd("PDEL", key, pattern)
	if err != nil {
		return err
	}

	return nil
}

// Expire sets or resets the timeout value on the requested entry.
func (db *Database) Expire(key string, id string, seconds int) (err error) {
	if db.pool == nil {
		return errUninitialized
	}

	_, err = db.runcmd("EXPIRE", key, id, strconv.Itoa(seconds))
	if err != nil {
		return err
	}

	return nil
}

// Persist removes the timeout value on the requested entry.
func (db *Database) Persist(key string, id string) (err error) {
	if db.pool == nil {
		return errUninitialized
	}

	_, err = db.runcmd("PERSIST", key, id)
	if err != nil {
		return err
	}

	return nil
}

// TTL returns the timeout value on the requested entry.
func (db *Database) TTL(key string, id string) (ttl float64, err error) {
	if db.pool == nil {
		return 0, errUninitialized
	}

	r, err := db.runcmd("TTL", key, id)
	if err != nil {
		return 0, err
	}

	return r.TTL, nil
}

// runcmd runs a command against the database.
func (db *Database) runcmd(cmd string, args ...string) (r *Response, err error) {
	if args == nil {
		return nil, errArgs
	}

	r = new(Response)

	err = db.pool.Do(radix.Cmd(r, cmd, args...))
	if err != nil {
		return nil, err
	}

	if !r.Ok {
		return nil, fmt.Errorf("%w: %s", errResponse, r.Err)
	}

	return r, nil
}

// connectJSON creates a connection and sets the output mode to JSON.
func connectJSON(net, addr string) (conn radix.Conn, err error) {
	conn, err = radix.Dial(net, addr)
	if err != nil {
		return nil, err
	}

	resp := new(Response)

	err = conn.Do(radix.Cmd(resp, "OUTPUT", "json"))
	if err != nil {
		conn.Close()

		return nil, err
	}

	if !resp.Ok {
		conn.Close()

		return nil, fmt.Errorf("%w: %s", errResponse, resp.Err)
	}

	return conn, nil
}
