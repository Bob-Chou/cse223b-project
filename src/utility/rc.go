package utility

import (
	"encoding/json"
	"fmt"
	"os"
	//"time"
)

const (
	DefaultPath = "nodes.rc"
)

type RC struct {
	Nodes []string
}

// Load runtime configuration files.
func LoadRC(p string) (*RC, error) {
	fin, e := os.Open(p)
	if e != nil {
		return nil, e
	}
	defer fin.Close()

	ret := new(RC)
	e = json.NewDecoder(fin).Decode(ret)
	if e != nil {
		return nil, e
	}

	return ret, nil
}

func (self *RC) Save(p string) error {
	fout, e := os.Create(p)
	if e != nil {
		return e
	}

	b, e := json.MarshalIndent(self, "", "    ")
	if e != nil {
		panic(e)
	}
	_, e = fout.Write(b)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintln(fout)
	if e != nil {
		return e
	}

	return fout.Close()
}