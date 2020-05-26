package test

import (
	"chord/ring"
	"log"
	"runtime/debug"
	"testing"
	"time"
)

func newNode(
	ip string,
	id uint64,
	join ring.Node,
	ready chan<- bool,
	catch chan<- error,
) *ring.Chord {
	// first Chord Node
	ch := ring.NewChord(ip)
	ch.ID = id
	if join == nil {
		ch.Create()
	} else {
		ch.Join(join)
	}

	go func() {
		if e := ch.Init(); e != nil {
			catch <- e
			return
		}
		if e := ch.Serve(ready); e != nil {
			catch <- e
			return
		}
	}()

	return ch
}

func ne(e error, t *testing.T) {
	if e != nil {
		debug.PrintStack()
		t.Fatal(e)
	}
}


func er(e error, t *testing.T) {
	if e == nil {
		debug.PrintStack()
		t.Fatal(e)
	}
}

func as(cond bool, t *testing.T) {
	if !cond {
		debug.PrintStack()
		t.Fatal("assertion failed")
	}
}

func TestNodeCreate(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	catch := make(chan error)
	ready := make(chan bool)
	ticker := time.NewTicker(10 * time.Second)

	// first Chord Node
	ch0 := newNode("localhost:1234", 0, nil, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// second Chord Node
	ch48 := newNode("localhost:1235", 48, ch0, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// third Chord Node
	ch18 := newNode("localhost:1236", 18, ch0, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// fourth Chord Node
	ch36 := newNode("localhost:1237", 36, ch18, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// wait for some time to see if Chord works without error
	select {
	case <-ticker.C:
		ticker.Stop()
	case e := <-catch:
		t.Fatal(e)
	}

	var next, prev ring.NodeInfo
	ne(ch0.Next(ch0.GetID(), &next), t)
	as(next.ID == ch18.GetID(), t)
	ne(ch0.Previous(ch0.GetID(), &prev), t)
	as(prev.ID == ch48.GetID(), t)
	ne(ch18.Next(ch18.GetID(), &next), t)
	as(next.ID == ch36.GetID(), t)
	ne(ch18.Previous(ch18.GetID(), &prev), t)
	as(prev.ID == ch0.GetID(), t)
	ne(ch36.Next(ch36.GetID(), &next), t)
	as(next.ID == ch48.GetID(), t)
	ne(ch36.Previous(ch36.GetID(), &prev), t)
	as(prev.ID == ch18.GetID(), t)
	ne(ch48.Next(ch48.GetID(), &next), t)
	as(next.ID == ch0.GetID(), t)
	ne(ch48.Previous(ch48.GetID(), &prev), t)
	as(prev.ID == ch36.GetID(), t)
}

func TestLookup(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	catch := make(chan error)
	ready := make(chan bool)
	ticker := time.NewTicker(10 * time.Second)

	// first Chord Node
	ch0 := newNode("localhost:1234", 0, nil, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// second Chord Node
	ch1 := newNode("localhost:1235", 48, ch0, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// third Chord Node
	ch2 := newNode("localhost:1236", 18, ch0, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// fourth Chord Node
	ch3 := newNode("localhost:1237", 36, ch2, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	log.Printf("wait for 20 seconds for chord setup")
	<-time.After(20*time.Second)

	var found ring.NodeInfo
	ne(ch0.FindSuccessor(1, &found), t)
	log.Printf("lookup %v, found %v", 1, found.ID)
	as(found.ID == 18, t)
	ne(ch1.FindSuccessor(1, &found), t)
	log.Printf("lookup %v, found %v", 1, found.ID)
	as(found.ID == 18, t)
	ne(ch2.FindSuccessor(1, &found), t)
	log.Printf("lookup %v, found %v", 1, found.ID)
	as(found.ID == 18, t)
	ne(ch3.FindSuccessor(1, &found), t)
	log.Printf("lookup %v, found %v", 1, found.ID)
	as(found.ID == 18, t)

	ne(ch0.FindSuccessor(63, &found), t)
	log.Printf("lookup %v, found %v", 63, found.ID)
	as(found.ID == 0, t)
	ne(ch1.FindSuccessor(63, &found), t)
	log.Printf("lookup %v, found %v", 63, found.ID)
	as(found.ID == 0, t)
	ne(ch2.FindSuccessor(63, &found), t)
	log.Printf("lookup %v, found %v", 63, found.ID)
	as(found.ID == 0, t)
	ne(ch3.FindSuccessor(63, &found), t)
	log.Printf("lookup %v, found %v", 63, found.ID)
	as(found.ID == 0, t)

	ne(ch0.FindSuccessor(37, &found), t)
	log.Printf("lookup %v, found %v", 37, found.ID)
	as(found.ID == 48, t)
	ne(ch1.FindSuccessor(37, &found), t)
	log.Printf("lookup %v, found %v", 37, found.ID)
	as(found.ID == 48, t)
	ne(ch2.FindSuccessor(37, &found), t)
	log.Printf("lookup %v, found %v", 37, found.ID)
	as(found.ID == 48, t)
	ne(ch3.FindSuccessor(37, &found), t)
	log.Printf("lookup %v, found %v", 37, found.ID)
	as(found.ID == 48, t)

	ne(ch0.FindSuccessor(19, &found), t)
	log.Printf("lookup %v, found %v", 19, found.ID)
	as(found.ID == 36, t)
	ne(ch1.FindSuccessor(19, &found), t)
	log.Printf("lookup %v, found %v", 19, found.ID)
	as(found.ID == 36, t)
	ne(ch2.FindSuccessor(19, &found), t)
	log.Printf("lookup %v, found %v", 19, found.ID)
	as(found.ID == 36, t)
	ne(ch3.FindSuccessor(19, &found), t)
	log.Printf("lookup %v, found %v", 19, found.ID)
	as(found.ID == 36, t)

	ne(ch0.FindSuccessor(0, &found), t)
	log.Printf("lookup %v, found %v", 0, found.ID)
	as(found.ID == 0, t)
	ne(ch1.FindSuccessor(0, &found), t)
	log.Printf("lookup %v, found %v", 0, found.ID)
	as(found.ID == 0, t)
	ne(ch2.FindSuccessor(0, &found), t)
	log.Printf("lookup %v, found %v", 0, found.ID)
	as(found.ID == 0, t)
	ne(ch3.FindSuccessor(0, &found), t)
	log.Printf("lookup %v, found %v", 0, found.ID)
	as(found.ID == 0, t)

	ne(ch0.FindSuccessor(18, &found), t)
	log.Printf("lookup %v, found %v", 18, found.ID)
	as(found.ID == 18, t)
	ne(ch1.FindSuccessor(18, &found), t)
	log.Printf("lookup %v, found %v", 18, found.ID)
	as(found.ID == 18, t)
	ne(ch2.FindSuccessor(18, &found), t)
	log.Printf("lookup %v, found %v", 18, found.ID)
	as(found.ID == 18, t)
	ne(ch3.FindSuccessor(18, &found), t)
	log.Printf("lookup %v, found %v", 18, found.ID)
	as(found.ID == 18, t)

	ne(ch0.FindSuccessor(36, &found), t)
	log.Printf("lookup %v, found %v", 36, found.ID)
	as(found.ID == 36, t)
	ne(ch1.FindSuccessor(36, &found), t)
	log.Printf("lookup %v, found %v", 36, found.ID)
	as(found.ID == 36, t)
	ne(ch2.FindSuccessor(36, &found), t)
	log.Printf("lookup %v, found %v", 36, found.ID)
	as(found.ID == 36, t)
	ne(ch3.FindSuccessor(36, &found), t)
	log.Printf("lookup %v, found %v", 36, found.ID)
	as(found.ID == 36, t)

	ne(ch0.FindSuccessor(48, &found), t)
	log.Printf("lookup %v, found %v", 48, found.ID)
	as(found.ID == 48, t)
	ne(ch1.FindSuccessor(48, &found), t)
	log.Printf("lookup %v, found %v", 48, found.ID)
	as(found.ID == 48, t)
	ne(ch2.FindSuccessor(48, &found), t)
	log.Printf("lookup %v, found %v", 48, found.ID)
	as(found.ID == 48, t)
	ne(ch3.FindSuccessor(48, &found), t)
	log.Printf("lookup %v, found %v", 48, found.ID)
	as(found.ID == 48, t)
}
