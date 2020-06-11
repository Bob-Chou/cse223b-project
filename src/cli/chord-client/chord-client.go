package main

import (
	"flag"
	"fmt"
	"os"
	"log"

	"chord/db"
	"chord/ring"
	"chord/hash"
	"utility"
)

var (
	usage = "usage: chord-client get `key` || chord-client set `key` `value`"
)

func noError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
func logError(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
}

// used for get
func getPara(args []string) string {
	if len(args) == 1 {
		return ""
	}
	return args[1]
}

// used for the parameter passed in set
func setPara(args []string) db.KV {
	if len(args) == 1 {
		return db.KV{K:"", V:""}
	} else if len(args) == 2 {
		return db.KV{K:args[1], V:""}
	}
	return db.KV{K:args[1], V:args[2]}
}

func keysPara(args []string) db.Pattern {
	if len(args) == 1 {
		return db.Pattern{Prefix: "", Suffix: ""}
	} else if len(args) == 2 {
		return db.Pattern{args[1],""}
	}
	return db.Pattern{Prefix:args[1], Suffix:args[2]}
}

func runCmd(args []string, ap string) bool {
	// the command for get / set / keys
	cmd := args[0]

	//ch := ring.NewChord(ap, "", db.NewStore())
	ch := ring.NewChordClient(ap, hash.EncodeKey(ap))
	switch cmd {
	case "get":
		var v string
		logError(ch.CGet(getPara(args), &v))
		fmt.Println(v)
	case "set":
		var ok bool
		logError(ch.CSet(setPara(args), &ok))
		fmt.Println(ok)
	}
	return false
}


func main() {
	// parse command line
	flag.Parse()

	// non-flag command-line arguments
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	/*
		1. loop through the rc back to find the first alive nodes
		2. run the cmd get / set
	*/

	// load the runtime configuration from `DefaultPath`
	rc, e := utility.LoadRC(utility.DefaultPath)
	noError(e)

	// first alive nodes as the access point
	ap := getAP(rc)

	fmt.Print(args)
	fmt.Print(ap)
	runCmd(args, ap)

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
