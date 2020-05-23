package db

import (
	"strings"
	"sync"
)

type KV struct {
	k string
	v string
}

type Pattern struct {
	prefix string
	suffix string
}

func (p *Pattern) Match(k string) bool {
	ret := strings.HasPrefix(k, p.prefix)
	ret = ret && strings.HasSuffix(k, p.suffix)
	return ret
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

// In-memory storage implementation. All calls always returns nil.
type Store struct {
	strs  map[string]string		// strs[key] = value

	strLock   sync.Mutex
}

func (self *Store) Get(k string, v *string) error {
	self.strLock.Lock()
	defer self.strLock.Unlock()

	*v = self.strs[k]

	return nil
}

func (self *Store) Set(kv KV, ok *bool) error {
	self.strLock.Lock()
	defer self.strLock.Unlock()

	if kv.v != "" {
		self.strs[kv.k] = kv.v
	} else {
		delete(self.strs, kv.k)
	}

	*ok = true

	return nil
}

func (self *Store) Keys(p Pattern, list *List) error {
	self.strLock.Lock()
	defer self.strLock.Unlock()

	ret := make([]string, 0, len(self.strs))

	for k := range self.strs {
		if p.Match(k) {
			ret = append(ret, k)
		}
	}

	list.L = ret

	return nil
}
