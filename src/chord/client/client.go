package client

import (
	"chord/db"
	"net/rpc"
	"net"
	"net/http"
)

type Client struct {
	addr string
}

// Key-value pair interfaces
func (self *Client) Get(k string, v *string) error {
	// connect to the server
	conn, e := rpc.DialHTTP("tcp", self.addr)
	if e != nil {
		return e
	}

	// perform the call
	e = conn.Call("Storage.Get", k, v)
	if e != nil {
		conn.Close()
		return e
	}

	// close the connection
	return conn.Close()
}

func (self *Client) Set(kv db.KV, ok *bool) error {
	// connect to the server
	conn, e := rpc.DialHTTP("tcp", self.addr)
	if e != nil {
		return e
	}

	// perform the call
	e = conn.Call("Storage.Set", kv, ok)
	if e != nil {
		conn.Close()
		return e
	}

	// close the connection
	return conn.Close()
}

func (self *Client) Keys(p db.Pattern, list *db.List) error {
	// connect to the server
	conn, e := rpc.DialHTTP("tcp", self.addr)
	if e != nil {
		return e
	}

	// reset the list for empty list
	list.L = []string{}

	// perform the call
	e = conn.Call("Storage.Keys", p, list)
	if e != nil {
		conn.Close()
		return e
	}

	// close the connection
	return conn.Close()
}

// Creates an RPC client that connects to addr.
func NewClient(addr string) db.Storage {
	return &Client{addr:addr}
}

// Serve as a backend based on the given configuration
func ServeBack(b *db.BackConfig) error {
	// Create an RPC server
	server := rpc.NewServer()

	// Register the Store member field in the b *trib.Config parameter under the name Storage
	server.RegisterName("Storage", b.Store)

	// Start an HTTP server
	l, e := net.Listen("tcp", b.Addr)

	// send a false when you encounter any error on starting your service.
	if e != nil {
		if b.Ready != nil {
			b.Ready <- false
		}
		return e
	}
	// Send a true over the Ready channel when the service is ready
	if b.Ready != nil {
		b.Ready <- true
	}

	return http.Serve(l, server)
}