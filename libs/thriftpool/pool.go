// Package pool implements a pool of net.Conn interfaces to manage and reuse them.
package thriftpool

import (
	"errors"
	"git.apache.org/thrift.git/lib/go/thrift"
)

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

// Conn represent a thrift connection.
type Conn struct {
	// Socket is the value returned by thrift.NewTSocket
	Socket thrift.TTransport
	// Client is the generic service client
	Client interface{}
}

// Pool interface describes a pool implementation. A pool should have maximum
// capacity. An ideal pool is threadsafe and easy to use.
type Pool interface {
	// Get returns a new connection from the pool. Closing the connections puts
	// it back to the Pool. Closing it when the pool is destroyed or full will
	// be counted as an error.
	Get() (*PoolConn, error)

	// Close closes the pool and all its connections. After Close() the pool is
	// no longer usable.
	Close()

	// Len returns the current number of connections of the pool.
	Len() int
}

func (c *Conn) Close() error {
	return c.Socket.Close()
}