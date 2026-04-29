// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package redis

import (
	"context"
	"errors"
	"time"
)

// Error represents an error returned in a command reply.
type Error string

func (err Error) Error() string { return string(err) }

// Conn represents a connection to a Redis server.
type Conn interface {
	// Close closes the connection.
	Close() error

	// Err returns a non-nil value when the connection is not usable.
	Err() error

	// Do sends a command to the server and returns the received reply.
	// This function will use the timeout which was set when the connection is created
	Do(commandName string, args ...interface{}) (reply interface{}, err error)

	// Send writes the command to the client's output buffer.
	Send(commandName string, args ...interface{}) error

	// Flush flushes the output buffer to the Redis server.
	Flush() error

	// Receive receives a single reply from the Redis server
	Receive() (reply interface{}, err error)
}

// Argument is the interface implemented by an object which wants to control how
// the object is converted to Redis bulk strings.
type Argument interface {
	// RedisArg returns a value to be encoded as a bulk string per the
	// conversions listed in the section 'Executing Commands'.
	// Implementations should typically return a []byte or string.
	RedisArg() interface{}
}

// Scanner is implemented by an object which wants to control its value is
// interpreted when read from Redis.
type Scanner interface {
	// RedisScan assigns a value from a Redis value. The argument src is one of
	// the reply types listed in the section `Executing Commands`.
	//
	// An error should be returned if the value cannot be stored without
	// loss of information.
	RedisScan(src interface{}) error
}

// ConnWithTimeout is an optional interface that allows the caller to override
// a connection's default read timeout. This interface is useful for executing
// the BLPOP, BRPOP, BRPOPLPUSH, XREAD and other commands that block at the
// server.
//
// A connection's default read timeout is set with the DialReadTimeout dial
// option. Applications should rely on the default timeout for commands that do
// not block at the server.
//
// All of the Conn implementations in this package satisfy the ConnWithTimeout
// interface.
//
// Use the DoWithTimeout and ReceiveWithTimeout helper functions to simplify
// use of this interface.
type ConnWithTimeout interface {
	Conn

	// DoWithTimeout sends a command to the server and returns the received reply.
	// The timeout overrides the readtimeout set when dialing the connection.
	DoWithTimeout(timeout time.Duration, commandName string, args ...interface{}) (reply interface{}, err error)

	// ReceiveWithTimeout receives a single reply from the Redis server.
	// The timeout overrides the readtimeout set when dialing the connection.
	ReceiveWithTimeout(timeout time.Duration) (reply interface{}, err error)
}

// ConnWithContext is an optional interface that allows the caller to control the command's life with context.
type ConnWithContext interface {
	Conn

	// DoContext sends a command to server and returns the received reply.
	// min(ctx,DialReadTimeout()) will be used as the deadline.
	// The connection will be closed if DialReadTimeout() timeout or ctx timeout or ctx canceled when this function is running.
	// DialReadTimeout() timeout return err can be checked by errors.Is(err, os.ErrDeadlineExceeded).
	// ctx timeout return err context.DeadlineExceeded.
	// ctx canceled return err context.Canceled.
	DoContext(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error)

	// ReceiveContext receives a single reply from the Redis server.
	// min(ctx,DialReadTimeout()) will be used as the deadline.
	// The connection will be closed if DialReadTimeout() timeout or ctx timeout or ctx canceled when this function is running.
	// DialReadTimeout() timeout return err can be checked by errors.Is(err, os.ErrDeadlineExceeded).
	// ctx timeout return err context.DeadlineExceeded.
	// ctx canceled return err context.Canceled.
	ReceiveContext(ctx context.Context) (reply interface{}, err error)
}

var errTimeoutNotSupported = errors.New("redis: connection does not support ConnWithTimeout")
var errContextNotSupported = errors.New("redis: connection does not support ConnWithContext")

// DoContext sends a command to server and returns the received reply.
// min(ctx,DialReadTimeout()) will be used as the deadline.
// The connection will be closed if DialReadTimeout() timeout or ctx timeout or ctx canceled when this function is running.
// DialReadTimeout() timeout return err can be checked by errors.Is(err, os.ErrDeadlineExceeded).
// ctx timeout return err context.DeadlineExceeded.
// ctx canceled return err context.Canceled.
func DoContext(c Conn, ctx context.Context, cmd string, args ...interface{}) (interface{}, error) {
	cwt, ok := c.(ConnWithContext)
	if !ok {
		return nil, errContextNotSupported
	}
	return cwt.DoContext(ctx, cmd, args...)
}

// DoWithTimeout executes a Redis command with the specified read timeout. If
// the connection does not satisfy the ConnWithTimeout interface, then an error
// is returned.
func DoWithTimeout(c Conn, timeout time.Duration, cmd string, args ...interface{}) (interface{}, error) {
	cwt, ok := c.(ConnWithTimeout)
	if !ok {
		return nil, errTimeoutNotSupported
	}
	return cwt.DoWithTimeout(timeout, cmd, args...)
}

// ReceiveContext receives a single reply from the Redis server.
// min(ctx,DialReadTimeout()) will be used as the deadline.
// The connection will be closed if DialReadTimeout() timeout or ctx timeout or ctx canceled when this function is running.
// DialReadTimeout() timeout return err can be checked by strings.Contains(e.Error(), "io/timeout").
// ctx timeout return err context.DeadlineExceeded.
// ctx canceled return err context.Canceled.
func ReceiveContext(c Conn, ctx context.Context) (interface{}, error) {
	cwt, ok := c.(ConnWithContext)
	if !ok {
		return nil, errContextNotSupported
	}
	return cwt.ReceiveContext(ctx)
}

// ReceiveWithTimeout receives a reply with the specified read timeout. If the
// connection does not satisfy the ConnWithTimeout interface, then an error is
// returned.
func ReceiveWithTimeout(c Conn, timeout time.Duration) (interface{}, error) {
	cwt, ok := c.(ConnWithTimeout)
	if !ok {
		return nil, errTimeoutNotSupported
	}
	return cwt.ReceiveWithTimeout(timeout)
}

// SlowLog represents a redis SlowLog
type SlowLog struct {
	// ID is a unique progressive identifier for every slow log entry.
	ID int64

	// Time is the unix timestamp at which the logged command was processed.
	Time time.Time

	// ExecutationTime is the amount of time needed for the command execution.
	ExecutionTime time.Duration

	// Args is the command name and arguments
	Args []string

	// ClientAddr is the client IP address (4.0 only).
	ClientAddr string

	// ClientName is the name set via the CLIENT SETNAME command (4.0 only).
	ClientName string
}

// Latency represents a redis LATENCY LATEST.
type Latency struct {
	// Name of the latest latency spike event.
	Name string

	// Time of the latest latency spike for the event.
	Time time.Time

	// Latest is the latest recorded latency for the named event.
	Latest time.Duration

	// Max is the maximum latency for the named event.
	Max time.Duration
}

// LatencyHistory represents a redis LATENCY HISTORY.
type LatencyHistory struct {
	// Time is the unix timestamp at which the event was processed.
	Time time.Time

	// ExecutationTime is the amount of time needed for the command execution.
	ExecutionTime time.Duration
}
