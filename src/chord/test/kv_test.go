package test

import (
	"chord/db"
	"chord/ring"
	"log"
	"testing"
	"time"
	"runtime/debug"
)

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

func newNode(
	ip string,
	id uint64,
	join string,
	ready chan<- bool,
	catch chan<- error,
) *ring.Chord {
	// first Chord Node
	ch := ring.NewChord(ip, join, db.NewStore())
	ch.ID = id

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

func TestNodeCreate(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	catch := make(chan error)
	ready := make(chan bool)
	ticker := time.NewTicker(10 * time.Second)

	// first Chord Node
	ch0 := newNode("localhost:1234", 0, "", ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// second Chord Node
	ch48 := newNode("localhost:1235", 48, ch0.IP, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// third Chord Node
	ch18 := newNode("localhost:1236", 18, ch0.IP, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// fourth Chord Node
	ch36 := newNode("localhost:1237", 36, ch18.IP, ready, catch)
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

	var kv db.KV
	var ok bool		 // return value for function`Set()`
	var val string   // return value for function `Get()`

	val = "_"
	ne(ch0.Get("", &val), t)
	as(val == "", t)

	kv = db.KV{K:"0", V:"0"}
	ch0.Set(kv, &ok)
	kv = db.KV{K:"10", V:"1"}
	ch0.Set(kv, &ok)
	kv = db.KV{K:"20", V:"2"}
	ch0.Set(kv, &ok)
	kv = db.KV{K:"30", V:"3"}
	ch0.Set(kv, &ok)
	kv = db.KV{K:"40", V:"4"}
	ch0.Set(kv, &ok)
	kv = db.KV{K:"50", V:"5"}
	ch0.Set(kv, &ok)
	kv = db.KV{K:"60", V:"6"}
	ch0.Set(kv, &ok)
	kv = db.KV{K:"70", V:"7"}
	ch0.Set(kv, &ok)
	kv = db.KV{K:"80", V:"8"}
	ch0.Set(kv, &ok)
	kv = db.KV{K:"90", V:"9"}
	ch0.Set(kv, &ok)

	val = "_"
	ne(ch0.Get("10", &val), t)
	as(val == "1", t)
	val = "_"
	ne(ch0.Get("20", &val), t)
	as(val == "2", t)
	val = "_"
	ne(ch0.Get("30", &val), t)
	as(val == "3", t)
	val = "_"
	ne(ch0.Get("40", &val), t)
	as(val == "4", t)
	val = "_"
	ne(ch0.Get("50", &val), t)
	as(val == "5", t)
	val = "_"
	ne(ch0.Get("60", &val), t)
	as(val == "6", t)
	val = "_"
	ne(ch0.Get("70", &val), t)
	as(val == "7", t)
	val = "_"
	ne(ch0.Get("80", &val), t)
	as(val == "8", t)
	val = "_"
	ne(ch0.Get("90", &val), t)
	as(val == "9", t)
}

