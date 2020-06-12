package test

import (
    "chord/db"
    "chord/ring"
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

func newNode(
    ip string,
    id uint64,
    join string,
    ready chan<- bool,
    catch chan<- error,
) *ring.Chord {
    // first Chord Node
    ch := ring.NewChord(ip, join, db.NewStore())
    ch.ID = id

    go func() {
        if e := ch.Init(); e != nil {
            catch <- e
            return
        }
        if e := ch.Serve(ready, nil); e != nil {
            catch <- e
            return
        }
    }()

    return ch
}

func newInteractNode(
    ip string,
    id uint64,
    join string,
    ready chan<- bool,
    catch chan<- error,
    kill <-chan bool,
) *ring.Chord {
    // first Chord Node
    ch := ring.NewChord(ip, join, db.NewStore())
    ch.ID = id

    go func() {
        if e := ch.Init(); e != nil {
            catch <- e
            return
        }
        if e := ch.Serve(ready, kill); e != nil {
            catch <- e
            return
        }
    }()

    return ch
}