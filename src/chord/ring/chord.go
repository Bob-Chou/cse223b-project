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
	"sync"
)

// dummy var to check if Chord implements NodeEntry interface
var _ NodeEntry = new(Chord)

// dummy var to check if Chord implements Node interface
var _ Node = new(Chord)

// dummy var to check if Chord implements db.Storage interface
//var _ db.Storage = new(Chord)

type Chord struct {
	// basic information
	NodeInfo
	// introducer IP address to Chord
	accessPoint Node
	// own client
	client *ChordClient
	// owner client
	owner *ChordClient
	// successor
	successor *ChordClient
	// predecessor
	predecessor *ChordClient
	// fingers list
	fingers []Node
	// r consecutive chord node for chain replication
	chain []Node
	// backend storage
	storage db.Storage
	// kv lock
	kvLock   sync.Mutex
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
    node.FindSuccessor(ch.ID, &found)
    ch.successor = NewChordClient(found.IP,found.ID)
}

// Stabilize is called periodically to verify n's immediate successor and tell
// the successor about n.
func(ch *Chord) Stabilize() {
	var x NodeInfo
	if e := ch.successor.Previous(ch.successor.ID, &x); e == nil {
		xclient:=NewChordClient(x.IP,x.ID)
		if In(x.ID,ch.ID,ch.successor.ID){
			//lock kv service
			ch.kvLock.Lock()
			var l db.List
			ch.successor.Keys(db.Pattern{"",""},&l)
			for i:= range l.L {
				k:=l.L[i]
				kid:=hash.EncodeKey(k)
				//need to migrate to x
				if RIn(kid,ch.ID,x.ID){
					log.Printf("Migrate key %v from node %v to node %v", kid,ch.successor.ID,x.ID)
					var v string
					ch.successor.Get(k,&v)
					var ok bool
					xclient.Set(db.KV{k,v},&ok)
				}
			}
			log.Printf("[%v] sets successor %v", ch.ID, x.ID)
			// close connection and free tcp file descriptor
			if ch.successor != nil {
				ch.successor.Reset()
			}
			ch.successor = xclient
			//unlock kv service
			ch.kvLock.Unlock()
		}
	} else if !ErrNotFound.Equals(e) {
		panic(e)
	}
	var ok bool
    ch.successor.Notify(&NodeInfo{ch.IP, ch.ID},&ok)
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

	if ch.fingers[i] == nil || found.ID != ch.fingers[i].GetID() {
		log.Printf("[%v] has new finger[%v] %v", ch.GetID(), i, found.ID)
		// close tcp connection and free file descriptor
		if ch.fingers[i] != nil {
			ch.fingers[i].(*ChordClient).Reset()
		}
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
	var nc NodeInfo
	if e := pre.FindSuccessor(pre.GetID(), &nc); e != nil {
		// TODO: Add node leave logic here
		log.Printf("[%v] %v (id %v) dies", ch.ID, pre.GetIP(), pre.GetID())
		// close tcp connection and free file descriptor
		if ch.predecessor != nil {
			ch.predecessor.Reset()
		}
		ch.predecessor = nil
	}
}

// PrecedingNode searches the local table for the highest predecessor of id
func(ch *Chord) PrecedingNode(id uint64) Node {
	id = id % uint64(1<<len(ch.fingers))

	for i := len(ch.fingers) - 1; i > 0; i-- {
		finger := ch.fingers[i]
		if finger != nil && In(finger.GetID(), ch.GetID(), id) {
			return finger
		}
	}

	if ch.successor != nil && In(ch.successor.GetID(), ch.GetID(), id) {
		return ch.successor
	}

	return nil
}

// get returns the value of specific key from underlying storage
func(ch *Chord) get(k string, v* string) error {
	return ch.storage.Get(k, v)
}

func(ch *Chord) set(kv db.KV, ok *bool) error {
	return ch.storage.Set(kv, ok)
}

func(ch *Chord) keys(p db.Pattern, list *db.List) error {
	return ch.storage.Keys(p, list)
}

// Get wraps the RPC interface db.Storage.Get
func(ch *Chord) Get(k string, v *string) error {
	ch.kvLock.Lock()
	defer ch.kvLock.Unlock()

	// find the key ID of key k
	Kid := hash.EncodeKey(k)
	//Kid = Kid % 72

	// find the owner of incoming id
	// throw error if owner is dead
	var nc NodeInfo
	if e := ch.FindSuccessor(Kid, &nc); e != nil {
		panic(fmt.Errorf("[%v] encounter error when finding owner node: %v", Kid, e))
	}
	ch.owner = NewChordClient(nc.IP, nc.ID)
	log.Printf("get key[%v] with kid[%v] in node[%v]\n", k, Kid, nc.ID)

	// complete get in owner node
	if e := ch.owner.Get(k, v); e != nil {
		panic(fmt.Errorf("[%v] encounter error when doing get: %v", k, e))
	}

	// close tcp connection and free file descriptor
	ch.owner.Reset()

	return nil
}

// Set wraps the RPC interface db.Storage.Set
func(ch *Chord) Set(kv db.KV, ok *bool) error {
	ch.kvLock.Lock()
	defer ch.kvLock.Unlock()

	// find the key ID of key k
	Kid := hash.EncodeKey(kv.K)
	//Kid = Kid % 72

	// find the owner of incoming id
	// throw error if owner is dead
	var nc NodeInfo
	if e := ch.FindSuccessor(Kid, &nc); e != nil {
		panic(fmt.Errorf("[%v] encounter error when finding owner node: %v", Kid, e))
	}
	ch.owner = NewChordClient(nc.IP, nc.ID)
	log.Printf("set key[%v] with kid[%v] in node[%v]\n", kv.K, Kid, nc.ID)

	// complete set in owner node
	if e := ch.owner.Set(kv, ok); e != nil {
		panic(fmt.Errorf("[%v] encounter error when doing set: %v", Kid, e))
	}

	return nil
}

// Keys wraps the RPC interface db.Storage.Keys
func(ch *Chord) Keys(p db.Pattern, list *db.List) error {
	return ch.storage.Keys(p, list)
}

func(ch *Chord) CGet(k string, v *string) error {
	return ch.Get(k, v)
}

func(ch *Chord) CSet(kv db.KV, ok *bool) error {
	return ch.Set(kv, ok)
}

// Notify wraps the RPC interface of NodeEntry.Notify
func(ch *Chord) Notify(node *NodeInfo, ok *bool) error {
	if (ch.predecessor == nil)||(In(node.ID,ch.predecessor.GetID(),ch.ID)){
		log.Printf("[%v] sets predecessor %v", ch.GetID(), node.ID)
		// close connection and free tcp file descriptor
		if ch.predecessor != nil {
			ch.predecessor.Reset()
		}
		ch.predecessor = NewChordClient(node.IP, node.ID)
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

	// Join the Chord
	if ch.accessPoint != nil {
		ch.Join(ch.accessPoint)
	} else {
		ch.Create()
	}

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
		chain:       make([]Node, 3),
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
