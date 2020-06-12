package test

import (
	"chord/db"
	"log"
	"testing"
	"time"
)

func TestKV(t *testing.T) {
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
	ne(ch48.Get("40", &val), t)
	as(val == "4", t)
	val = "_"
	ne(ch48.Get("50", &val), t)
	as(val == "5", t)
	val = "_"
	ne(ch18.Get("60", &val), t)
	as(val == "6", t)
	val = "_"
	ne(ch18.Get("70", &val), t)
	as(val == "7", t)
	val = "_"
	ne(ch36.Get("80", &val), t)
	as(val == "8", t)
	val = "_"
	ne(ch36.Get("90", &val), t)
	as(val == "9", t)
}

func TestMigration(t *testing.T) {
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


	// wait for some time to see if Chord works without error
	select {
	case <-ticker.C:
		ticker.Stop()
	case e := <-catch:
		t.Fatal(e)
	}

	var kv db.KV
	var ok bool		 // return value for function`Set()`
	var val string   // return value for function `Get()`

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

	ticker2 := time.NewTicker(10 * time.Second)
	select {
	case <-ticker2.C:
		ticker2.Stop()
	case e := <-catch:
		t.Fatal(e)
	}

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
	ne(ch48.Get("40", &val), t)
	as(val == "4", t)
	val = "_"
	ne(ch48.Get("50", &val), t)
	as(val == "5", t)
	val = "_"
	ne(ch18.Get("60", &val), t)
	as(val == "6", t)
	val = "_"
	ne(ch18.Get("70", &val), t)
	as(val == "7", t)
	val = "_"
	ne(ch36.Get("80", &val), t)
	as(val == "8", t)
	val = "_"
	ne(ch36.Get("90", &val), t)
	as(val == "9", t)

}