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

// Get has two steps
// 1.find element from dict
// 2.move element to the end of queue (we assume Front is the end)
func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		// move ele to the end of queue
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry) // conversion: interface{} -> entry
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		// remove from ll
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// delete from dict
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	// if key exist in cache, update the value and move the element to the end of queue.
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// if nBytes is excess maxBytes, remove the oldest element
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) GetNBytes() int64 {
	return c.nbytes
}

func (c *Cache) GetMaxBytes() int64 {
	return c.maxBytes
}

func (c *Cache) Len() int {
	return len(c.cache)
}
