package db

type KV struct {
	k string
	v string
}

type Pattern struct {
	prefix string
	suffix string
}

type List struct {
	L []string
}

type Storage interface {
	// Get returns the value of the given key
	Get(k string, v *string) error
	// Set sets the value of the given key, and returns if succeed
	Set(kv KV, ok *bool) error
	// Keys returns the keys matched the given pattern
	Keys(p Pattern, list *List) error
}