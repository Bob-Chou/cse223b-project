package test

import (
	"chord/ring"
	"log"
	"testing"
	"time"
	"fmt"
	"strconv"
	"math/rand"
)

func TestHop(t *testing.T) {
    //var nodeIDs [8]uint64
    var pathlength [8]int

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	catch := make(chan error)
	ready := make(chan bool)
	ticker := time.NewTicker(60 * time.Second)

    var i,n int
    var ch0 *ring.Chord

    n = 40

    i = 0
	// first Chord Node
	ch0 = newNode("127.0.0.1:1234", 0, "", ready, catch)
	//nodeIDs[i] = ch0.ID
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Fatal("service not ready, timed out")
	case <-ready:
	case e := <-catch:
		t.Fatal(e)
	}
    for i=1;i<n;i++{
		newNode("127.0.0.1:"+strconv.Itoa(1234+i), 0, ch0.IP, ready, catch)
		select {
		case <-ticker.C:
			ticker.Stop()
			t.Fatal("service not ready, timed out")
		case <-ready:
		case e := <-catch:
			t.Fatal(e)
		}
	}

	// wait for some time to see if Chord works without error
	select {
	case <-ticker.C:
		log.Printf("Chord ring is formed successfully.")
		ticker.Stop()
	case e := <-catch:
		t.Fatal(e)
	}

    var nc ring.CountHop
    var id ring.HopIn

    var sum,min,max int
    sum = 0
    min = 1000000
    max = 0
    for i:=0;i<100;i++{
    	id.ID = uint64(rand.Intn(256*256-1)) //0-63
    	id.Count = 1
    	nc.Count = 1
		if e := ch0.FindSuccessor(id, &nc); e != nil {
			panic(fmt.Errorf("[%v] encounter error when finding owner node: %v", id.ID, e))
		}
		//log.Printf("The successor for [%v] is [%v]",id.ID,nc.ID)
		//log.Printf("The hop count to find [%v] is [%v]",id.ID,nc.Count)
		sum += nc.Count
		if nc.Count < min {min = nc.Count}
		if nc.Count > max {max = nc.Count}
		pathlength[nc.Count]+=1
    }
    log.Printf("nodes: %v, hop count total: %v, min: %v, max: %v",n, sum, min, max)
    for i:=0;i<8;i++{
    	log.Printf("[%v]hops: %v\n", i, pathlength[i])
    }

}