package ring

import (
	"chord/db"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strings"
	"sync"
	"time"
)

// dummy var to check if Chord implements NodeEntry interface
var _ NodeEntry = new(Chord)

// dummy var to check if Chord implements Node interface
var _ Node = new(Chord)

// dummy var to check if Chord implements db.Storage interface
var _ db.Storage = new(Chord)

type Chord struct {
	// basic information
	NodeInfo
	// mutex for fingers list
	fingersMtx sync.RWMutex
	// fingers list
	fingers []Node
	// r consecutive chord node for chain replication
	chain []Node
}

func(ch *Chord) SetFinger(i int, n Node) {
	ch.fingersMtx.Lock()
	defer ch.fingersMtx.Unlock()
	ch.fingers[i] = n
}

func(ch *Chord) GetFinger(i int) Node {
	ch.fingersMtx.RLock()
	defer ch.fingersMtx.RUnlock()
	return ch.fingers[i]
}

// GetID wraps Node.GetID
func(ch *Chord) GetID() uint64 {
	return ch.ID
}

// GetIP wraps Node.GetIP
func(ch *Chord) GetIP() string {
	return ch.IP
}

// Create creates a new Chord ring
func(ch *Chord) Create() {
	panic("todo")
}

// Join joins a Chord ring containing the given node
func(ch *Chord) Join(node *NodeEntry) {
	panic("todo")
}

// Stabilize is called periodically to verify n's immediate successor and tell
// the successor about n.
func(ch *Chord) Stabilize() {
	panic("todo")
}

// FixFingers is called periodically to refresh finger table entries
func(ch *Chord) FixFingers() {
	for i := 1; i < len(ch.fingers); i++ {
		var found NodeInfo
		if e := ch.FindSuccessor(ch.ID+uint64(2^(i-1)), &found); e != nil {
			panic(fmt.Errorf("encounter error when fix fingers: %v", e))
		}

		if ch.GetFinger(i) == nil || found.ID != ch.GetFinger(i).GetID() {
			ch.SetFinger(i, NewChordClient(found.IP, found.ID))
		}
	}
}

// CheckPredecessor is called periodically to check whether predecessor has
// failed
func(ch *Chord) CheckPredecessor() {
	if ch.GetFinger(0) == nil {
		return
	}
	p := ch.GetFinger(0)
	var nc NodeInfo
	if p.FindSuccessor(p.GetID(), &nc) != nil {
		// TODO: Add node leave logic here
		log.Printf("%v (id %v) dies", p.GetIP(), p.GetID())
		ch.SetFinger(0, nil)
	}
}

// PrecedingNode searches the local table for the highest predecessor of id
func (ch *Chord) PrecedingNode(id uint64) Node {
	id = id % uint64(2^(len(ch.fingers)-1))

	for i := len(ch.fingers) - 1; i > 0; i-- {
		if Between(id, ch.ID, ch.GetFinger(i).GetID()) {
			return ch.GetFinger(i)
		}
	}

	return nil
}

// Get wraps the RPC interface db.Storage.Get
func(ch *Chord) Get(k string, v *string) error {
	panic("todo")
}

// Set wraps the RPC interface db.Storage.Set
func(ch *Chord) Set(kv db.KV, ok *bool) error {
	panic("todo")
}

// Keys wraps the RPC interface db.Storage.Keys
func(ch *Chord) Keys(p db.Pattern, list *db.List) error {
	panic("todo")
}

// Notify wraps the RPC interface of NodeEntry.Notify
func(ch *Chord) Notify(node *NodeInfo, ok *bool) error {
	panic("todo")
}

// FindSuccessor wraps the RPC interface of NodeEntry.FindSuccessor
func(ch *Chord) FindSuccessor(id uint64, found *NodeInfo) error {
	id = id % uint64(2^(len(ch.fingers)-1))

	successor := ch.GetFinger(1)
	if Between(id, ch.ID, successor.GetID()) {
		*found = NodeInfo{
			IP: successor.GetIP(),
			ID: successor.GetID(),
		}
		return nil
	}

	return ch.PrecedingNode(id).FindSuccessor(id, found)
}

// Next wraps the RPC interface of NodeEntry.Next
func(ch *Chord) Next(id uint64, found *NodeInfo) error {

	if id != ch.ID {
		return fmt.Errorf("wrong node id")
	}

	if ch.GetFinger(1) == nil {
		return fmt.Errorf("nil successor")
	}

	successor := ch.GetFinger(1)
	*found = NodeInfo{
		IP: successor.GetIP(),
		ID: successor.GetID(),
	}

	return nil
}

// Next wraps the RPC interface of NodeEntry.Previous
func(ch *Chord) Previous(id uint64, found *NodeInfo) error {

	if id != ch.ID {
		return fmt.Errorf("wrong node id")
	}

	if ch.GetFinger(0) == nil {
		return fmt.Errorf("nil predecessor")
	}

	successor := ch.GetFinger(0)
	*found = NodeInfo{
		IP: successor.GetIP(),
		ID: successor.GetID(),
	}

	return nil
}

// Init completes all preparation logic, when it returns, it should be ready to
// provide KV service. It also returns the encountered error if any
func(ch *Chord) Init() error {
	return nil
}

// Serve is a blocking call to serve, it never returns.  It also returns the
// encountered error if any
func(ch *Chord) Serve() error {
	catch := make(chan error)
	quit := make(chan bool)

	// Chord RPC services
	go func() {
		chordRPC := NodeEntry(ch)

		addrSplit := strings.Split(ch.IP, ":")
		port := addrSplit[len(addrSplit)-1]

		e := rpc.RegisterName(port+"/NodeEntry", chordRPC)
		if e != nil {
			catch <- e
		}

		l, e := net.Listen("tcp", ch.IP)

		if e != nil {
			catch <- e
			return
		}

		for {
			conn, e := l.Accept()
			if e != nil {
				catch <- e
				return
			}
			go rpc.ServeConn(conn)
		}
	}()

	go func() {
		<-time.After(time.Second)
		ch.FixFingers()
	}()

	go func() {
		<-time.After(time.Second)
		ch.Stabilize()
	}()

	e := <-catch
	close(quit)

	return e
}

func Between(id, selfID, succID uint64) bool {
	if succID > selfID {
		return id <= succID && id > selfID
	} else {
		return id <= succID || id > selfID
	}
}
