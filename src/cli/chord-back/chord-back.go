package main

import (
	"flag"
	"log"
	"strconv"
	"time"

	"utility"
	"chord/client"
	"chord/db"
)
const (
	DefaultPath = "backs.rc"
)

func noError(e error) {
	if e != nil {
		log.Printf("error is ", e)
	}
}
func run(addr string) {
	backConfig := db.BackConfig {
		Addr: addr,
		Store: db.NewStore(),
	}

	log.Printf("back-end serving on %s", backConfig.Addr)
	noError(client.ServeBack(&backConfig))

}

func main() {
	// parse command line
	flag.Parse()
	args := flag.Args()		// non-flag command-line arguments


	// load the runtime configuration from `DefaultPath`
	rc, e := utility.LoadRC(DefaultPath)
	noError(e)

	if len(args) == 0 {
		for _, addr := range rc.Backs {
			go run(addr)
			time.Sleep(1*time.Second)
		}
	} else {
		for _, a := range args {
			i, e := strconv.Atoi(a)
			noError(e)

			go run(rc.Backs[i])
		}
	}
	select {}

}