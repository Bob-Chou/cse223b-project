package main

import (
	"flag"
	"fmt"
	"os"
	"log"
	"strconv"

	"chord/db"
	"chord/ring"
	"chord/hash"
	"utility"
)

var (
	usage = "usage: \n" +
		"\tchord-client get `key` || chord-client set `key` `value`\n" +
		"\tchord-client i get `key` || chord-client i set `key` `value`"
)

func noError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func checkAP(ap string) error {
	client := ring.NewChordClient(ap, hash.EncodeKey(ap))
	if e := client.Dial(); e != nil {
		return fmt.Errorf("[%v] node is not ready!", ap)
	}
	return nil
}

// getAP uses to get the access point of nodes
func getAP(rc *utility.RC) string {
	for _, ip := range rc.Nodes {
		if checkAP(ip) == nil {
			return ip
		}
	}
	return ""
}

// getPara is used to generate a parameter in `get`
func getPara(args []string) string {
	if len(args) == 1 {
		return ""
	}
	return args[1]
}

// setPara is used to generate a parameter in `set`
func setPara(args []string) db.KV {
	if len(args) == 1 {
		return db.KV{K:"", V:""}
	} else if len(args) == 2 {
		return db.KV{K:args[1], V:""}
	}
	return db.KV{K:args[1], V:args[2]}
}

func runCmd(args []string, ap string) bool {
	// access point of chord
	ch := ring.NewChordClient(ap, hash.EncodeKey(ap))

	// the command for get / set
	cmd := args[0]
	switch cmd {
	case "get":
		var v string
		noError(ch.CGet(getPara(args), &v))
		fmt.Println(v)
	case "set":
		var ok bool
		noError(ch.CSet(setPara(args), &ok))
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

	// load the runtime configuration from `DefaultPath`
	rc, e := utility.LoadRC(utility.DefaultPath)
	noError(e)

	i, e := strconv.Atoi(args[0])
	// usage 1 or usage 2
	if e != nil {
		// first alive nodes as the access point
		ap := getAP(rc)
		runCmd(args, ap)
	} else {
		ap:= rc.Nodes[i]
		noError(checkAP(ap))
		runCmd(args[1:], ap)
	}
}
