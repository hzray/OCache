package lru

import (
	"container/list"
)

// Cache 缓存对象，定义了缓存的基本结构
type Cache struct {
	maxBytes    int64                         // max memory size that can be taken
	nbytes      int64                         // amount of memory have been taken
	cacheLL     *list.List                    // 管理元素顺序(实现LRU)
	cache       map[string]*list.Element      // store the relationship between key and value
	onEvicted   func(key string, value Value) // optional and executed when an entry is purged.
	K           int                           // LRU-K
	historyLL   *list.List                    // 管理历史记录
	history     map[string]*list.Element      // 存放历史记录键值对
	historyRest int
}

// entry cache linked list存储的结构体
// advantage to store key value in linked list: when merge element, use key to delete relationship from dict
type entry struct {
	key   string
	value Value
}

// Value use Len() to count how many bytes it takes
type Value interface {
	Len() int
}

// historyCounter history linked list存储的结构体
type historyCounter struct {
	key  string
	time int
}

// New is the Constructor of Cache
func New(k int, maxBytes int64, historyMax int, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:    maxBytes,
		cacheLL:     list.New(),
		cache:       make(map[string]*list.Element),
		onEvicted:   onEvicted,
		K:           k,
		historyLL:   list.New(),
		history:     make(map[string]*list.Element),
		historyRest: historyMax,
	}
}

// removeOldest remove oldest element from cache
func (c *Cache) removeOldest() {
	ele := c.cacheLL.Back()
	if ele != nil {
		// remove from cacheLL
		c.cacheLL.Remove(ele)
		kv := ele.Value.(*entry)
		// delete from dict
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

// addToCache add key value pair from history to cache
func (c *Cache) addToCache(key string, value Value) {
	// if key exist in cache, update the value and move the element to the end of queue.
	if ele, ok := c.cache[key]; ok {
		c.cacheLL.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.cacheLL.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// if nBytes is excess maxBytes, remove the oldest element
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.removeOldest()
	}
}

func (c *Cache) historyReplacingCheck() {
	if c.historyRest == 0 {
		// true: trigger LRU history
		tail := c.historyLL.Back()
		c.historyLL.Remove(tail)
		delete(c.history, tail.Value.(historyCounter).key)
		c.historyRest++
	}
}

func (c *Cache) deleteFromHistory(key string) {
	if ele, ok := c.history[key]; !ok {
		return
	} else {
		delete(c.history, key)
		c.historyLL.Remove(ele)
		c.historyRest++
	}
}

// Get has two steps
// 1.find element from dict
// 2.move element to the end of queue (we assume Front is the end)
func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		// move ele to the end of queue
		c.cacheLL.MoveToFront(ele)
		kv := ele.Value.(*entry) // conversion: interface{} -> entry
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) Add(key string, value Value) {
	var (
		hc historyCounter
	)

	// special case: LRU
	if c.K == 1 {
		c.addToCache(key, value)
		return
	}

	// missed in cache, then check history to incr visited time
	hEle, ok := c.history[key]
	if ok {
		// auto-incr visited count
		hc = hEle.Value.(historyCounter)
		hc.time++

		if hc.time >= c.K {
			// true: removed from history, and add into cache
			c.deleteFromHistory(key)
			c.addToCache(key, value)
		}
		// write back to history
		hEle.Value = hc
	} else {
		c.historyReplacingCheck()
		hc = historyCounter{key: key, time: 1}
		hEle = c.historyLL.PushFront(hc)
		c.historyRest--
	}
	c.history[key] = hEle
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
