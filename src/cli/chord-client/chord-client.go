package main

import (
	"flag"
	"fmt"
	"os"

	"chord/db"
	"chord/client"
)

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

func runCmd(storage db.Storage, args []string) bool {
	// the command for get / set / keys
	cmd := args[0]
	switch cmd {
	case "get":
		var v string
		logError(storage.Get(getPara(args), &v))
		fmt.Println(v)
	case "set":
		var ok bool
		logError(storage.Set(setPara(args), &ok))
		fmt.Println(ok)
	case "keys":
		var keys db.List
		pattern :=keysPara(args)
		storage.Keys(pattern, &keys)
		//logError()
		for _, key := range keys.L {
			fmt.Println(key)
		}
	}
	return false
}


func main() {
	// parse command line
	flag.Parse()

	// non-flag command-line arguments
	args := flag.Args()
	if len(args) < 1 {
		// TODO add usage
		//fmt.Fprintln(os.Stderr, help)
		os.Exit(1)
	}

	addr := args[0]
	storage := client.NewClient(addr)

	cmdArgs := args[1:]
	if len(cmdArgs) == 0 {
		// TODO add usage
		fmt.Println()
	} else {
		runCmd(storage, cmdArgs)
	}
}