package test

import (
    "runtime/debug"
    "testing"
)

func ne(e error, t *testing.T) {
    if e != nil {
        debug.PrintStack()
        t.Fatal(e)
    }
}


func er(e error, t *testing.T) {
    if e == nil {
        debug.PrintStack()
        t.Fatal(e)
    }
}

func as(cond bool, t *testing.T) {
    if !cond {
        debug.PrintStack()
        t.Fatal("assertion failed")
    }
}