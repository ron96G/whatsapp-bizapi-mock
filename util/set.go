package util

import (
	"sync"
)

type Set struct {
	l map[string]interface{}
	m sync.RWMutex
}

func NewSet() *Set {
	return &Set{
		l: map[string]interface{}{},
		m: sync.RWMutex{},
	}
}

func (x *Set) Add(key string) {
	x.m.Lock()
	x.l[key] = struct{}{}
	x.m.Unlock()
}

func (x *Set) Del(key string) {
	x.m.Lock()
	delete(x.l, key)
	x.m.Unlock()
}

func (x *Set) Contains(key string) bool {
	x.m.RLock()
	defer x.m.RUnlock()
	_, found := x.l[key]
	return found
}
