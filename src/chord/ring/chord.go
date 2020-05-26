package ring

import (
	"chord/db"
	"chord/hash"
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
	// successor
	successor Node
	// predecessor
	predecessor Node
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
func(ch *Chord) FixFingers(i int) {
	if i < 0 || i >= len(ch.fingers) {
		panic(fmt.Sprintf("[%v] FixFingers index error", ch.ID))
	}

	var found NodeInfo
	if e := ch.FindSuccessor(ch.GetID()+uint64(1<<i), &found); e != nil {
		// TODO: handle nodes leaving case here
		panic(fmt.Errorf("[%v] encounter error when fix fingers: %v", ch.ID, e))
	}

	if ch.GetFinger(i) == nil || found.ID != ch.GetFinger(i).GetID() {
		log.Printf("[%v] has new finger[%v] %v", ch.GetID(), i, found.ID)
		ch.SetFinger(i, NewChordClient(found.IP, found.ID))
	}
}

// CheckPredecessor is called periodically to check whether predecessor has
// failed
func(ch *Chord) CheckPredecessor() {
	if ch.predecessor == nil {
		return
	}
	pre := ch.predecessor
	var nc NodeInfo
	if e := pre.FindSuccessor(pre.GetID(), &nc); e != nil {
		// TODO: Add node leave logic here
		log.Printf("[%v] %v (id %v) dies", ch.ID, pre.GetIP(), pre.GetID())
		ch.predecessor = nil
	}
}

// PrecedingNode searches the local table for the highest predecessor of id
func(ch *Chord) PrecedingNode(id uint64) Node {
	id = id % uint64(1<<len(ch.fingers))

	for i := len(ch.fingers) - 1; i >= 0; i-- {
		finger := ch.GetFinger(i)
		if finger != nil && In(finger.GetID(), ch.GetID(), id) {
			return finger
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
	if (ch.predecessor == nil)||((node.ID > ch.predecessor)&&(node.ID<ch.ID)){
		ch.predecessor = node
	}
	return nil
}

// FindSuccessor wraps the RPC interface of NodeEntry.FindSuccessor
func(ch *Chord) FindSuccessor(id uint64, found *NodeInfo) error {
	id = id % (1 << len(ch.fingers))
	//log.Printf("[%v] start to find successor of %v", ch.ID, id)

	if RIn(id, ch.GetID(), ch.successor.GetID()) {
		*found = NodeInfo{
			IP: ch.successor.GetIP(),
			ID: ch.successor.GetID(),
		}
		//log.Printf("[%v] successor of %v has been found: %v", ch.ID, id, found.ID)
		return nil
	}

	redirect := ch.PrecedingNode(id)
	if redirect == nil {
		return ErrNotReady
	}

	//log.Printf("[%v] redirect FindSuccessor(%v) to %v", ch.ID, id, redirect.GetID())
	var ans NodeInfo
	if e := redirect.FindSuccessor(id, &ans); e !=nil {
		return e
	}
	*found = ans

	return nil
}

// Next wraps the RPC interface of NodeEntry.Next
func(ch *Chord) Next(id uint64, found *NodeInfo) error {

	if id != ch.GetID() {
		return ErrWrongID
	}

	if ch.successor == nil {
		return ErrNotFound
	}

	*found = NodeInfo{
		IP: ch.successor.GetIP(),
		ID: ch.successor.GetID(),
	}

	return nil
}

// Next wraps the RPC interface of NodeEntry.Previous
func(ch *Chord) Previous(id uint64, found *NodeInfo) error {

	if id != ch.GetID() {
		log.Printf("[%v] wrong id %v", ch.ID, id)
		return ErrWrongID
	}

	if ch.predecessor == nil {
		return ErrNotFound
	}

	pre := ch.predecessor
	*found = NodeInfo{
		IP: pre.GetIP(),
		ID: pre.GetID(),
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
func(ch *Chord) Serve(ready chan<- bool) error {
	catch := make(chan error)
	quit := make(chan bool)
	done := make(chan bool, 1)
	// Chord RPC services
	go func() {
		chordServer := &ChordServer{ch}
		addrSplit := strings.Split(ch.IP, ":")
		port := addrSplit[len(addrSplit)-1]

		e := rpc.RegisterName(port+"/NodeEntry", chordServer)
		if e != nil {
			catch <- e
			return
		}

		l, e := net.Listen("tcp", ch.IP)

		if e != nil {
			catch <- e
			return
		}

		done <- true

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
		ticker := time.NewTicker(100 * time.Millisecond)
		next := 1
		for {
			next = next%len(ch.fingers)
			select {
			case <-quit:
				ticker.Stop()
				return
			case <-ticker.C:
				ch.FixFingers(next)
			}
			next++
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-quit:
				ticker.Stop()
				return
			case <-ticker.C:
				ch.Stabilize()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-quit:
				ticker.Stop()
				return
			case <-ticker.C:
				ch.CheckPredecessor()
			}
		}
	}()

	<-done
	log.Printf("[%v] service ready", ch.ID)
	ready <- true
	e := <-catch
	close(quit)

	return e
}

// NewChord create a new chord node and return the pointer to it
func NewChord(ip string) *Chord {
	id := hash.EncodeKey(ip)
	ch := Chord{
		NodeInfo:   NodeInfo{ip, id},
		fingersMtx: sync.RWMutex{},
		fingers:    make([]Node, hash.MaxHashBits),
		chain:      make([]Node, 3),
	}
	return &ch
}

// In determines whether a given ID locates in (nid1, nid2)
func In(id, nid1, nid2 uint64) bool {
	if nid2 > nid1 {
		return id < nid2 && id > nid1
	} else {
		return id < nid2 || id > nid1
	}
}

// RIn determines whether a given ID locates in (nid1, nid2]
func RIn(id, nid1, nid2 uint64) bool {
	if nid2 > nid1 {
		return id <= nid2 && id > nid1
	} else {
		return id <= nid2 || id > nid1
	}
}
