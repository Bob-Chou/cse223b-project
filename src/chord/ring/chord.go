package ring

import (
	"chord/db"
	"sync"
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
	//predecessor and successor
	predecessor Node
	successor Node
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
	ch.predecessor = nil
	ch.successor = ch.ID
}

// Join joins a Chord ring containing the given node
func(ch *Chord) Join(node *NodeEntry) {
	ch.predecessor = nil
    ch.successor = node.FindSuccessor(ch.ID)
}

// Stabilize is called periodically to verify n's immediate successor and tell
// the successor about n.
func(ch *Chord) Stabilize() {
	x:=ch.successor.predecessor
	if x.ID > ch.ID && x.ID < ch.successor.ID{
		ch.successor = x
	}
    ch.successor.Notify(ch)
}

// FixFingers is called periodically to refresh finger table entries
func(ch *Chord) FixFingers() {
	panic("todo")
}

// CheckPredecessor is called periodically to check whether predecessor has
// failed
func(ch *Chord) CheckPredecessor() {
	panic("todo")
}

// PrecedingNode searches the local table for the highest predecessor of id
func(ch *Chord) PrecedingNode(id uint64) Node {
	panic("todo")
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
	if (ch.predecessor == nil)||((node.ID > ch.predecessor)&&(node.ID<ch.ID)){
		ch.predecessor = node
	}
	return nil
}

// FindSuccessor wraps the RPC interface of NodeEntry.FindSuccessor
func(ch *Chord) FindSuccessor(id uint64, found *NodeInfo) error {
	panic("todo")
}

// Init completes all preparation logic, when it returns, it should be ready to
// provide KV service. It also returns the encountered error if any
func(ch *Chord) Init() error {
	return nil
}

// Serve is a blocking call to serve, it never returns.  It also returns the
// encountered error if any
func(ch *Chord) Serve() error {
	return nil
}