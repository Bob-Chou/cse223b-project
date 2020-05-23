package ring

import (
	"net/rpc"
	"strings"
	"sync"
)

var _ NodeEntry = new(ChordClient)

type ChordClient struct {
	NodeInfo
	conn     *rpc.Client
	connLock sync.Mutex
}

func (c *ChordClient) dial() error {
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

func (c *ChordClient) reset() {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = nil
}

func (c *ChordClient) call(call string, args interface{}, ret interface{}) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	e := c.conn.Call(call, args, ret)
	return e
}

func (c *ChordClient) rpc(name string, args interface{}, ret interface{}) error {
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
func(c *ChordClient) FindSuccessor(id uint64, found *NodeInfo) error {
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

// NewChordClient returns a pointer to a ChordClient
func NewChordClient(ip string, id uint64) *ChordClient {
	return &ChordClient{
		NodeInfo: NodeInfo{IP: ip, ID: id},
		conn:     nil,
		connLock: sync.Mutex{},
	}
}