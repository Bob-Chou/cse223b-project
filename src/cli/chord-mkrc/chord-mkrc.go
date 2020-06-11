package main

import (
	"flag"
	"strconv"
	"log"

	"utility"
)

const (
	MaxNodes 	= 1 << 16
	Hostname 	= "localhost"
)

var (
	nnodes = flag.Int("nnodes", 1, "number of nodes")		// flag nnodes with default value 1
)

func main() {
	// parse command line
	flag.Parse()
	// check the number of flag options
	if flag.NFlag() != 1 {
		flag.Usage()
	}

	// check if the number of nodes is valid
	if *nnodes > MaxNodes {
		log.Fatal("too many nodes")
	}

	// create the runtime configuration
	rc := utility.RC{Nodes: make([]string, *nnodes)}

	// create `nnodes` port starting from `port`
	port := utility.RandPort()
	for i := 0; i < *nnodes; i++ {
		rc.Nodes[i] = Hostname + ":" + strconv.Itoa(port)
		port++
	}

	// save the configuration into the file `DefaultPath`
	if e := rc.Save(utility.DefaultPath); e != nil {
		log.Fatal(e)
	}
}