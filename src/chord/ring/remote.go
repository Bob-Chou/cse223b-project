package ring

import (
	"chord/db"
	"net/rpc"
	"strings"
	"sync"
)

type ChordClient struct {
	NodeInfo
	conn     *rpc.Client
	connLock sync.Mutex
}

func(c *ChordClient) dial() error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if c.conn == nil {
		conn, e := rpc.Dial("tcp", c.IP)
		if e != nil {
			return e
		}
		c.conn = conn
	}
	return nil
}

func(c *ChordClient) reset() {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = nil
}

func(c *ChordClient) call(call string, args interface{}, ret interface{}) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	e := c.conn.Call(call, args, ret)
	return e
}


func(c *ChordClient) rpc(name string, args interface{}, ret interface{}) error {
	if e := c.dial(); e != nil {
		return e
	}

	if e := c.call(name, args, ret); e != nil {
		// reconnect
		c.reset()
		if e := c.dial(); e != nil {
			return e
		}

		if e := c.call(name, args, ret); e != nil {
			c.reset()
			return e
		}
	}

	return nil
}

// Dial uses to check if the node is able to serve
func(c *ChordClient) Dial() error {
	return c.dial()
}

// Reset help close tcp connection and free file descriptor
func(c *ChordClient) Reset() {
	c.reset()
}

// GetID wraps Node.GetID
func(c *ChordClient) GetID() uint64 {
	return c.ID
}

// GetIP wraps Node.GetIP
func(c *ChordClient) GetIP() string {
	return c.IP
}

// Notify wraps the RPC interface of NodeEntry.Notify
func(c *ChordClient) Notify(node *NodeInfo, ok *bool) error {
	addrSplit := strings.Split(c.IP, ":")
	port := addrSplit[len(addrSplit)-1]
	name := port + "/NodeEntry.Notify"
	return c.rpc(name, node, ok)
}

// FindSuccessor wraps the RPC interface of NodeEntry.FindSuccessor
func(c *ChordClient) FindSuccessor(id HopIn, found *CountHop) error {
	addrSplit := strings.Split(c.IP, ":")
	port := addrSplit[len(addrSplit)-1]
	name := port + "/NodeEntry.FindSuccessor"
	return c.rpc(name, id, found)
}

// Next wraps the RPC interface of NodeEntry.Next
func(c *ChordClient) Next(id uint64, found *NodeInfo) error {
	addrSplit := strings.Split(c.IP, ":")
	port := addrSplit[len(addrSplit)-1]
	name := port + "/NodeEntry.Next"
	return c.rpc(name, id, found)
}

// Previous wraps the RPC interface of NodeEntry.Previous
func(c *ChordClient) Previous(id uint64, found *NodeInfo) error {
	addrSplit := strings.Split(c.IP, ":")
	port := addrSplit[len(addrSplit)-1]
	name := port + "/NodeEntry.Previous"
	return c.rpc(name, id, found)
}

// Get wraps the RPC interface of NodeEntry.Get
func(c *ChordClient) Get(k string, v *string) error {
	addrSplit := strings.Split(c.IP, ":")
	port := addrSplit[len(addrSplit)-1]
	name := port + "/NodeEntry.Get"
	return c.rpc(name, k, v)
}

func(c *ChordClient) Set(kv db.KV, ok *bool) error {
	addrSplit := strings.Split(c.IP, ":")
	port := addrSplit[len(addrSplit)-1]
	name := port + "/NodeEntry.Set"
	return c.rpc(name, kv, ok)
}

// Get wraps the RPC interface of NodeEntry.Get
func(c *ChordClient) CGet(k string, v *string) error {
	addrSplit := strings.Split(c.IP, ":")
	port := addrSplit[len(addrSplit)-1]
	name := port + "/NodeEntry.CGet"
	return c.rpc(name, k, v)
}

func(c *ChordClient) CSet(kv db.KV, ok *bool) error {
	addrSplit := strings.Split(c.IP, ":")
	port := addrSplit[len(addrSplit)-1]
	name := port + "/NodeEntry.CSet"
	return c.rpc(name, kv, ok)
}

// NewChordClient returns a pointer to a ChordClient
func NewChordClient(ip string, id uint64) *ChordClient {
	return &ChordClient{
		NodeInfo: NodeInfo{IP: ip, ID: id},
		conn:     nil,
		connLock: sync.Mutex{},
	}
}

func(c *ChordClient) Keys(p db.Pattern, list *db.List) error{
	addrSplit := strings.Split(c.IP, ":")
	port := addrSplit[len(addrSplit)-1]
	name := port + "/NodeEntry.Keys"
	return c.rpc(name, p, list)
}

type ChordServer struct {
	entry *Chord
}

// FindSuccessor is called to find the successor of a given id
func(c *ChordServer) FindSuccessor(id HopIn, found *CountHop) error {
	return c.entry.FindSuccessor(id, found)
}

// Notify is called when the given node thinks it might be our predecessor
func(c *ChordServer) Notify(node *NodeInfo, ok *bool) error {
	return c.entry.Notify(node, ok)
}

// Next returns the successor, or returns error if has no successor
func(c *ChordServer) Next(id uint64, next *NodeInfo) error {
	return c.entry.Next(id, next)
}

// Previous returns the predecessor, or returns error if has no predecessor
func(c *ChordServer) Previous(id uint64, prev *NodeInfo) error {
	return c.entry.Previous(id, prev)
}

func(c *ChordServer) Get(k string, v *string) error {
  return c.entry.get(k, v)
}

func(c *ChordServer) Set(kv db.KV, ok *bool) error {
	return c.entry.set(kv, ok)
}

func(c *ChordServer) Keys(p db.Pattern, list *db.List) error {
	return c.entry.keys(p,list)
}

func(c *ChordServer) CGet(k string, v *string) error {
	return c.entry.CGet(k, v)
}

func(c *ChordServer) CSet(kv db.KV, ok *bool) error {
	return c.entry.CSet(kv, ok)
}

var _ NodeEntry = new(ChordClient)
var _ NodeEntry = new(ChordServer)
