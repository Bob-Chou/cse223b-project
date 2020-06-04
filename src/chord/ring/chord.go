package ring

import (
	"chord/db"
	"chord/hash"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strings"
	"time"
)

const (
	TimeOut  = 500 * time.Millisecond
	NumBacks = 3
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
	// introducer IP address to Chord
	accessPoint Node
	// successor
	successor *ChordClient
	// predecessor
	predecessor *ChordClient
	// fingers list
	fingers []Node
	// r consecutive chord nodes
	successors []Node
	// backend storage
	storage db.Storage
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
	ch.successor = NewChordClient(ch.GetIP(),ch.GetID())
}

// Join joins a Chord ring containing the given node
func(ch *Chord) Join(node Node) {
	ch.predecessor = nil
	var found NodeInfo
    if e := node.FindSuccessor(ch.ID, &found); e != nil {
    	panic(e)
	}
    ch.successor = NewChordClient(found.IP,found.ID)
}

// Stabilize is called periodically to verify n's immediate successor and tell
// the successor about n.
func(ch *Chord) Stabilize() {
	next := 0
	for {
		// fix successor
		if !ch.Ping(ch.successor) {
			if ch.successors[next] != nil {
				log.Printf("[%v] update successor %v", ch.GetID(), ch.successors[next].GetID())
				ch.successor = ch.successors[next].(*ChordClient)
			}
			next++
			if next >= len(ch.successors) {
				break
			}
			continue
		}

		// update successor list
		ch.FixSuccessorList()

		var x NodeInfo
		if e := ch.successor.Previous(ch.successor.GetID(), &x); e == nil {
			if In(x.ID,ch.ID,ch.successor.GetID()){
				log.Printf("[%v] sets successor %v", ch.GetID(), x.ID)
				ch.successor = NewChordClient(x.IP,x.ID)
			}
		} else if !ErrNotFound.Equals(e) {
			next = 0
			continue
		}

		var ok bool
		if ch.successor.Notify(&NodeInfo{ch.IP, ch.ID},&ok) != nil {
			next = 0
			continue
		}

		return
	}

	// Cannot fix ring because all backup successors are dead
	panic(ErrBrokenService)
}

// FixFingers is called periodically to refresh finger table entries
func(ch *Chord) FixFingers(i int) {
	if i < 0 || i >= len(ch.fingers) {
		panic(fmt.Sprintf("[%v] FixFingers index error", ch.ID))
	}

	var found NodeInfo
	if e := ch.FindSuccessor(ch.GetID()+uint64(1<<i), &found); e != nil {
		panic(e)
	}

	if ch.fingers[i] == nil || found.ID != ch.fingers[i].GetID() {
		log.Printf("[%v] has new finger[%v] %v", ch.GetID(), i, found.ID)
		ch.fingers[i] = NewChordClient(found.IP, found.ID)
	}
}

// CheckPredecessor is called periodically to check whether predecessor has
// failed
func(ch *Chord) CheckPredecessor() {
	if ch.predecessor == nil {
		return
	}
	pre := ch.predecessor
	if !ch.Ping(pre) {
		log.Printf("[%v] %v (id %v) dies", ch.ID, pre.GetIP(), pre.GetID())
		ch.predecessor = nil
	}
}

// PrecedingNode searches the local table for the highest predecessor of id
func(ch *Chord) PrecedingNode(id uint64) Node {
	id = id % uint64(1<<len(ch.fingers))

	for i := len(ch.fingers) - 1; i > 0; i-- {
		finger := ch.fingers[i]
		if finger != nil && In(finger.GetID(), ch.GetID(), id) {
			if ch.Ping(finger) {
				return finger
			}
		}
	}

	// check backup list
	for i := len(ch.successors) - 1; i >= 0; i-- {
		if ch.successors[i] != nil && In(ch.successors[i].GetID(), ch.GetID(), id) {
			if ch.Ping(ch.successors[i]) {
				return ch.successors[i]
			}
		}
	}

	// check direct successor
	if ch.successor != nil && In(ch.successor.GetID(), ch.GetID(), id) {
		if ch.Ping(ch.successor) {
			return  ch.successor
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
	if (ch.predecessor == nil)||(In(node.ID,ch.predecessor.GetID(),ch.ID)){
		log.Printf("[%v] sets predecessor %v", ch.GetID(), node.ID)
		ch.predecessor = NewChordClient(node.IP, node.ID)
	}
	return nil
}

// FindSuccessor wraps the RPC interface of NodeEntry.FindSuccessor
func(ch *Chord) FindSuccessor(id uint64, found *NodeInfo) error {
	id = id % (1 << len(ch.fingers))
	//log.Printf("[%v] start to find successor of %v", ch.ID, id)

	// check direct successor
	if RIn(id, ch.GetID(), ch.successor.GetID()) {
		if ch.Ping(ch.successor) {
			*found = NodeInfo{
				IP: ch.successor.GetIP(),
				ID: ch.successor.GetID(),
			}
			return nil
		}
	}

	// check backup list
	for _, s := range ch.successors {
		if s != nil && RIn(id, ch.GetID(), s.GetID()) {
			if ch.Ping(s) {
				*found = NodeInfo{
					IP: s.GetIP(),
					ID: s.GetID(),
				}
				return nil
			}
		}
	}

	//log.Printf("[%v] redirect FindSuccessor(%v) to %v", ch.ID, id, redirect.GetID())
	var ans NodeInfo
	for {
		redirect := ch.PrecedingNode(id)
		if redirect == nil {
			return ErrBrokenService
		}

		// returned node might have been dead when forward request to it
		// therefore we need retry logic here
		if e := redirect.FindSuccessor(id, &ans); e != nil {
			if ErrBrokenService.Equals(e) {
				return e
			}
			continue
		}

		break
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

// Previous wraps the RPC interface of NodeEntry.Previous
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

// Ping test whether a Chord node is dead
func(ch *Chord) Ping(n Node) bool {
	rtn := make(chan error)
	go func() {
		var nc NodeInfo
		rtn <- n.Next(n.GetID(), &nc)
	}()

	select {
	case e := <-rtn:
		return e == nil || ErrNotFound.Equals(e) || ErrWrongID.Equals(e)
	case <-time.After(TimeOut):
		return false
	}
}

// FixSuccessorList updates successors by walking from the direct successor and
// fetch for all backup successors
func(ch *Chord) FixSuccessorList() {
	var cur Node
	cur = ch.successor

	for i := 0; i < len(ch.successors); i++ {
		var next NodeInfo
		if cur != nil && cur.Next(cur.GetID(), &next) == nil {
			if ch.successors[i] == nil || next.ID != ch.successors[i].GetID() {
				log.Printf("[%v] sets successor list @%v: %v", ch.GetID(), i, next.ID)
				ch.successors[i] = NewChordClient(next.IP, next.ID)
			}
		}
		cur = ch.successors[i]
	}

	return
}

// Successors returns the successor list of current node
func(ch *Chord) Successors() []Node {
	return append([]Node{ch.successor}, ch.successors...)
}

// Init completes all preparation logic, when it returns, it should be ready to
// provide KV service. It also returns the encountered error if any
func(ch *Chord) Init() error {
	return nil
}

// Serve is a blocking call to serve, it never returns.  It also returns the
// encountered error if any
func(ch *Chord) Serve(ready chan<- bool, kill <-chan bool) error {
	catch := make(chan error)
	quit := make(chan bool)
	done := make(chan bool, 1)

	// Join the Chord
	if ch.accessPoint != nil {
		ch.Join(ch.accessPoint)
	} else {
		ch.Create()
	}

	// Chord RPC services
	go func() {
		barrier := make(chan bool)
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

		// handle incoming RPC calls
		go func() {
			for {
				conn, e := l.Accept()
				if e != nil {
					select {
					case <-quit:
					default:
						catch <- e
					}
					return
				}
				go func() {
					done := make(chan bool, 1)
					go func() {
						rpc.ServeConn(conn)
						close(done)
					}()
					go func() {
						select {
						case <-done:
						case <-quit:
							conn.Close()
						}
					}()
					<-done
				}()
			}
		}()

		// close RPC listener if receive quit signal
		go func() {
			<-quit
			if e := l.Close(); e != nil {
				log.Println(e)
				catch <- e
			}
			close(barrier)
		}()

		<-barrier
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

	select {
	case <-kill:
		close(quit)
		log.Printf("[%v] is killed by outside", ch.ID)
		return nil
	case e := <-catch:
		close(quit)
		return e
	}
}

// NewChord create a new chord node and return the pointer to it
func NewChord(ip, accessIP string, storage db.Storage) *Chord {
	id := hash.EncodeKey(ip)

	var ap Node
	if accessIP == "" || accessIP == ip {
		ap = nil
	} else {
		ap = NewChordClient(accessIP, hash.EncodeKey(accessIP))
	}
	ch := Chord{
		NodeInfo:    NodeInfo{ip, id},
		accessPoint: ap,
		fingers:     make([]Node, hash.MaxHashBits),
		successors:  make([]Node, NumBacks),
		storage:     storage,
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
