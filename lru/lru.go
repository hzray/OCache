package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         // max memory size that can be taken
	nbytes    int64                         // amount of memory have been taken
	ll        *list.List                    // all values will be put into ll
	cache     map[string]*list.Element      // store the relationship between key and value
	onEvicted func(key string, value Value) // optional and executed when an entry is purged.

}

// the data type list.Element's Value field
// advantage to store key value in linked list: when merge element, use key to delete relationship from dict
type entry struct {
	key   string
	value Value
}

// Value use Len() to count how many bytes it takes
type Value interface {
	Len() int
}

// New is the Constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}
