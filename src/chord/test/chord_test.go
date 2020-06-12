package test

import (
	"chord/hash"
	"chord/ring"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

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
}

func TestLookup(t *testing.T) {
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
	ch1 := newNode("localhost:1235", 48, ch0.IP, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// third Chord Node
	ch2 := newNode("localhost:1236", 18, ch0.IP, ready, catch)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	// fourth Chord Node
	ch3 := newNode("localhost:1237", 36, ch2.IP, ready, catch)
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

	ne(ch0.FindSuccessor(1<<hash.MaxHashBits-1, &found), t)
	log.Printf("lookup %v, found %v", 1<<hash.MaxHashBits-1, found.ID)
	as(found.ID == 0, t)
	ne(ch1.FindSuccessor(1<<hash.MaxHashBits-1, &found), t)
	log.Printf("lookup %v, found %v", 1<<hash.MaxHashBits-1, found.ID)
	as(found.ID == 0, t)
	ne(ch2.FindSuccessor(1<<hash.MaxHashBits-1, &found), t)
	log.Printf("lookup %v, found %v", 1<<hash.MaxHashBits-1, found.ID)
	as(found.ID == 0, t)
	ne(ch3.FindSuccessor(1<<hash.MaxHashBits-1, &found), t)
	log.Printf("lookup %v, found %v", 1<<hash.MaxHashBits-1, found.ID)
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

func TestSuccessorList(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	chords := make([]*ring.Chord, ring.NumBacks)
	catch := make(chan error)
	ready := make(chan bool)
	ticker := time.NewTicker(10 * time.Second)

	chords[0] = newNode("localhost:1234", 0, "", ready, catch)
	<-ready

	for i := 1; i < len(chords); i++ {
		ip := fmt.Sprintf("localhost:%v", 1234+i)
		chords[i] = newNode(ip, uint64(i), chords[0].IP, ready, catch)
	}

	done := make(chan bool, 1)
	go func() {
		for i := 1; i < len(chords); i++ {
			<-ready
		}
		done <- true
	}()
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-done:
	case e := <-catch:
		t.Fatal(e)
	}

	<-time.After(20*time.Second)

	for i := 0; i < len(chords); i++ {
		ch := chords[i]
		nodes := ch.Successors()
		for j, node := range nodes {
			t.Logf("[%v] successors[%v] is %v", ch.ID, j, node.GetID())
			as(node.GetID() == uint64((i + j + 1) % len(chords)), t)
		}
	}
}

func TestKillChord(t *testing.T) {
	catch := make(chan error)
	ready := make(chan bool)
	kill := make(chan bool, 1)
	ticker := time.NewTicker(10 * time.Second)

	ch0 := newInteractNode("localhost:5678", 376, "", ready, catch, kill)
	client := ring.NewChordClient(ch0.IP, ch0.ID)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}

	var found ring.NodeInfo
	ne(client.FindSuccessor(0, &found), t)

	kill <- true
	<-time.After(1 * time.Second)

	e := client.FindSuccessor(0, &found)
	er(e, t)
	as(!ch0.Ping(client), t)
}

func TestFixSuccessor(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	chords := make([]*ring.Chord, ring.NumBacks+1)
	catch := make(chan error)
	ready := make(chan bool)
	ticker := time.NewTicker(10 * time.Second)

	signals := make([]chan bool, len(chords))
	for i := 0; i < len(signals); i++ {
		signals[i] = make(chan bool, 1)
	}

	chords[0] = newInteractNode("localhost:1234", 0, "", ready, catch, signals[0])
	<-ready

	for i := 1; i < len(chords); i++ {
		ip := fmt.Sprintf("localhost:%v", 1234+i)
		chords[i] = newInteractNode(ip, uint64(i), chords[0].IP, ready, catch, signals[i])
	}

	done := make(chan bool, 1)
	go func() {
		for i := 1; i < len(chords); i++ {
			<-ready
		}
		done <- true
	}()
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-done:
	case e := <-catch:
		t.Fatal(e)
	}

	<-time.After(15*time.Second)

	for i := 0; i < len(chords); i++ {
		ch := chords[i]
		nodes := ch.Successors()
		for j, node := range nodes {
			t.Logf("[%v] successors[%v] is %v", ch.ID, j, node.GetID())
			as(node.GetID() == uint64((i + j + 1) % len(chords)), t)
		}
	}

	signals[len(signals)-1] <- true
	signals[len(signals)-2] <- true

	<-time.After(5*time.Second)

	for i := 0; i < len(chords)-2; i++ {
		ch := chords[i]
		nodes := ch.Successors()
		for j, node := range nodes {
			t.Logf("[%v] successors[%v] is %v", ch.ID, j, node.GetID())
			as(node.GetID() == uint64((i + j + 1) % (len(chords)-2)), t)
		}
	}

	for i := 0; i < len(chords)-2; i++ {
		var prev, next ring.NodeInfo
		ch := chords[i]

		// check ring integrity
		var ans uint64
		ne(ch.Next(ch.ID, &next), t)
		ans = uint64((i+1)%(len(chords)-2))
		t.Logf("[%v]'s successor should be %v, get %v", ch.ID, ans, next.ID)
		as(next.ID == ans, t)
		ne(ch.Previous(ch.ID, &prev), t)
		ans = uint64((i-1+(len(chords)-2))%(len(chords)-2))
		t.Logf("[%v]'s predecessor should be %v, get %v", ch.ID, ans, prev.ID)
		as(prev.ID == ans, t)

		// check functionality correctness
		for j := 0; j < len(chords)-2; j++ {
			var found ring.NodeInfo
			ne(ch.FindSuccessor(uint64(j), &found), t)
			t.Logf("[%v] lookup %v should be %v, get %v", ch.ID, j, j, found.ID)
			as(found.ID == uint64(j), t)
		}
	}
}
