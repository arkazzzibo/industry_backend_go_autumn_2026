package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}
type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	mu       sync.Mutex
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		ll:       list.List{},
		items:    make(map[K]*list.Element),
	}
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.ll.MoveToFront(elem)
		return elem.Value.(entry[K, V]).value, true
	}
	return value, ok
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		val := elem.Value.(entry[K, V])
		val.value = value
		elem.Value = val // присваиваем чтобы не потерять тк копия
		return
	}

	if c.capacity == c.ll.Len() {
		oldest := c.ll.Back()
		oldestEntry := oldest.Value.(entry[K, V])

		delete(c.items, oldestEntry.key)
		c.ll.Remove(oldest)
	}

	c.items[key] = c.ll.PushFront(entry[K, V]{
		key:   key,
		value: value,
	})
	return
}
