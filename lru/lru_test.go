package lru

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	t.Run("k=1, unlimited max bytes", func(t *testing.T) {
		lru := New(1, int64(0), 30, nil)
		lru.Add("key1", String("1234"))
		if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
			t.Fatalf("cache hit key1=1234 failed")
		}
		if _, ok := lru.Get("key2"); ok {
			t.Fatalf("cache miss key2 failed")
		}
	})

	t.Run("k=1, limited max bytes", func(t *testing.T) {
		lru := New(1, int64(20), 30, nil)
		lru.Add("key1", String("1234"))
		lru.Add("key2", String("1234"))
		lru.Add("key3", String("1234"))
		if _, ok := lru.Get("key1"); ok {
			t.Fatalf("cache miss key1 failed")
		}
		if v, ok := lru.Get("key2"); !ok || string(v.(String)) != "1234" {
			t.Fatalf("cache hit key2=1234 failed")
		}
	})

	t.Run("k=2, unlimited max bytes", func(t *testing.T) {
		lru := New(2, int64(0), 30, nil)
		lru.Add("key1", String("1234"))
		lru.Add("key1", String("1234"))
		lru.Add("key3", String("1234"))
		if _, ok := lru.Get("key3"); ok {
			t.Fatalf("cache miss key1 failed")
		}
		if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
			t.Fatalf("cache hit key2=1234 failed")
		}
	})

	t.Run("k=2, limited max bytes", func(t *testing.T) {
		lru := New(2, int64(20), 30, nil)
		lru.Add("key1", String("1234"))
		lru.Add("key1", String("1234"))
		lru.Add("key2", String("1234"))
		lru.Add("key2", String("1234"))
		lru.Add("key3", String("1234"))
		lru.Add("key3", String("1234"))

		if _, ok := lru.Get("key1"); ok {
			t.Fatalf("cache miss key1 failed")
		}
		if v, ok := lru.Get("key2"); !ok || string(v.(String)) != "1234" {
			t.Fatalf("cache hit key2=1234 failed")
		}
	})
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(1, int64(cap), 30, nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get(k1); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(1, int64(10), 30, callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))
	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestDeleteHistory(t *testing.T) {
	t.Run("k=2, unlimited max bytes", func(t *testing.T) {
		lru := New(2, int64(0), 2, nil)
		lru.Add("key1", String("1234"))
		lru.Add("key2", String("1234"))
		lru.Add("key3", String("1234"))
		lru.Add("key1", String("1234"))
		lru.Add("key3", String("1234"))

		if _, ok := lru.Get("key1"); ok {
			t.Fatalf("cache miss key1 failed")
		}
		if v, ok := lru.Get("key3"); !ok || string(v.(String)) != "1234" {
			t.Fatalf("cache hit key2=1234 failed")
		}
	})
}
