package util

import (
	"fmt"
	"sync"
)

type LockedList struct {
	l []interface{}
	m sync.RWMutex
}

func NewLockedList() *LockedList {
	return &LockedList{
		l: []interface{}{},
		m: sync.RWMutex{},
	}
}

func (x *LockedList) Add(item interface{}) {
	x.m.Lock()
	x.l = append(x.l, item)
	x.m.Unlock()
}

func (x *LockedList) Del(item interface{}) {
	x.m.Lock()
	fmt.Printf("Trying to delete item %s\n", item)
	for i, elem := range x.l {
		if elem == item {
			fmt.Printf("Removing item %s\n", item)
			x.l = remove(x.l, i)
			break
		}
	}
	fmt.Printf("Deletion complete %v\n", x.l)
	x.m.Unlock()
}

func (x *LockedList) Contains(item interface{}) bool {
	x.m.RLock()
	defer x.m.RUnlock()
	return contains(x.l, item)
}

func contains(slice []interface{}, item interface{}) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

func remove(s []interface{}, i int) []interface{} {
	s[i] = s[len(s)-1]
	s[len(s)-1] = ""
	return s[:len(s)-1]
}
