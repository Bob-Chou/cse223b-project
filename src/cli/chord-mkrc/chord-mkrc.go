package main

import (
	"flag"
	"strconv"
	"log"

	"utility"
)

const (
	MaxBacks 	= 256	//TODO
	Hostname 	= "localhost"
	DefaultPath = "backs.rc"
)

var (
	nback = flag.Int("nbacks", 1, "number of back-ends")		// flag nbacks with default value 1
)

func main() {
	// parse command line
	flag.Parse()
	// check the number of flag options
	if flag.NFlag() != 1 {
		flag.Usage()
	}

	// check if the number of back-ends is valid
	if *nback > MaxBacks {
		log.Fatal("too many backs")
	}

	// create the runtime configuration
	rc := utility.RC{Backs: make([]string, *nback)}

	// create `nbacks` port starting from `port`
	port := utility.RandPort()
	for i := 0; i < *nback; i++ {
		rc.Backs[i] = Hostname + ":" + strconv.Itoa(port)
		port++
	}

	// save the configuration into the file `DefaultPath`
	if e := rc.Save(DefaultPath); e != nil {
		log.Fatal(e)
	}
}