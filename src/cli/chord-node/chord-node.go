package main

import (
	"flag"
	"log"
	"strconv"
	"time"

	"utility"
	"chord/db"
	"chord/ring"
	"chord/hash"
)

func noError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func run(ip, ap string) {

	ch := ring.NewChord(ip, ap, db.NewStore())
	go func() {
		if e := ch.Init(); e != nil {
			log.Fatal("service not ready, timed out")
		}

		ready := make(chan bool)
		kill := make(chan bool)
		if e := ch.Serve(ready, kill); e != nil {
			log.Fatal(e)
		}
	}()
}

func main() {
	// parse command line
	flag.Parse()
	args := flag.Args()		// non-flag command-line arguments


	// load the runtime configuration from `DefaultPath`
	rc, e := utility.LoadRC(utility.DefaultPath)
	noError(e)

	if len(args) == 0 {
		// scan for address on this machine
		ap := ""
		for _, ip := range rc.Nodes {
			go run(ip, ap)
			ap = ip
			time.Sleep(1*time.Second)
		}
	} else {
		ap := getAP(rc)
		for n, a := range args {
			i, e := strconv.Atoi(a)
			noError(e)
			if n != 0 {
				for ap == "" {
					<-time.After(500 * time.Millisecond)
					ap = getAP(rc)
				}
			}
			go run(rc.Nodes[i], ap)
		}
	}
	select {}

}

func getAP(rc *utility.RC) string {
	for _, ip := range rc.Nodes {
		client := ring.NewChordClient(ip, hash.EncodeKey(ip))
		if e := client.Dial(); e == nil {
			return ip
		}
	}
	return ""
}