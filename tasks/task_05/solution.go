package main

type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		capacity: capacity,
		items:    make(map[K]V),
	}
}
func (c *Cache[K, V]) Get(k K) (v V, ok bool) {
	item, ok := c.items[k]
	return item, ok
}
func (c *Cache[K, V]) Set(k K, v V) bool {
	if c.capacity <= 0 {
		return false
	}

	if _, ok := c.items[k]; ok {
		c.items[k] = v
		return true
	}

	if len(c.items) >= c.capacity {
		return false
	}

	c.items[k] = v
	return true
}
