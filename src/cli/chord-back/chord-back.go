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

var (
	//frc       = flag.String("rc", trib.DefaultRCPath, "bin storage config file")
	//verbose   = flag.Bool("v", false, "verbose logging")
	//readyAddr = flag.String("ready", "", "ready notification address")
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

	//if *readyAddr != "" {
	//	backConfig.Ready = utility.Chan(*readyAddr, backConfig.Addr)
	//}

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
}