package test

import (
	"sort"
	"testing"

	"chord/db"
)

func kv(k, v string) db.KV {
	return db.KV{K:k, V:v}
}

func pat(p, s string) db.Pattern{
	return db.Pattern{Prefix:p, Suffix:s}
}

func TestStore(t *testing.T) {

	var val string   // return value for function `Get()`
	var ok bool      // return value for function`Set()`
	var keys db.List // return value for function `Keys()`

	// initialize the store
	store := db.NewStore()

	val = "_"
	ne(store.Get("", &val), t)
	as(val == "", t)

	val = "_"
	ne(store.Get("hello", &val), t)
	as(val == "", t)

	ne(store.Set(kv("k1", "v1"), &ok), t)
	as(ok, t)
	val = "_"
	ne(store.Get("k1", &val), t)
	as(val == "v1", t)

	ne(store.Set(kv("k2", "v2"), &ok), t)
	as(ok, t)
	val = "_"
	ne(store.Get("k2", &val), t)
	as(val == "v2", t)

	ne(store.Keys(pat("k", ""), &keys), t)
	sort.Strings(keys.L)
	as(len(keys.L) == 2, t)
	as(keys.L[0] == "k1", t)
	as(keys.L[1] == "k2", t)
}
