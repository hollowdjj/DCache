package lru

import (
	"container/list"
	"time"
)

//a LRU cache with TTL
type LRUCache struct {
	//list and hash map
	list  *list.List
	cache map[interface{}]*list.Element

	//callback when element is deleted
	OnDroped func(key interface{}, val interface{})
}

//value of list node
type Entry struct {
	Key      interface{}
	Val      interface{}
	ExpireAt time.Time
}

//creat a new LRU cache
func New() *LRUCache {
	return &LRUCache{
		list:  list.New(),
		cache: make(map[interface{}]*list.Element),
	}
}

//add element to cache
func (l *LRUCache) Add(key interface{}, val interface{}, ttl time.Duration) {
	//lazy initialization
	if l.cache == nil {
		l.cache = make(map[interface{}]*list.Element)
		l.list = list.New()
	}

	//if key exists, update and move to head
	node, ok := l.cache[key]
	if ok {
		l.list.MoveToFront(node)
		nodeEntry := node.Value.(*Entry)
		nodeEntry.Val = val
		nodeEntry.ExpireAt = time.Now().Add(ttl)
		return
	}

	//add new element
	newNode := l.list.PushFront(&Entry{key, val, time.Now().Add(ttl)})
	l.cache[key] = newNode
}

//look up cache according to key
func (l *LRUCache) Get(key interface{}) (node *Entry, ok bool) {
	if l.cache == nil {
		return
	}
	if node, hit := l.cache[key]; hit {
		l.list.MoveToFront(node)
		return node.Value.(*Entry), true
	}
	return
}

//delete cache according key
func (l *LRUCache) Del(key interface{}) {
	if l.cache == nil {
		return
	}
	if _, hit := l.cache[key]; hit {
		l.removeCache(key)
	}
}

//remove least recently used cache
func (l *LRUCache) RemoveLeastUsed() {
	if l.cache == nil {
		return
	}
	target := l.list.Back()
	if target != nil {
		l.list.Remove(target)
		key := target.Value.(*Entry).Key
		delete(l.cache, key)
	}
}

//remove cache
func (l *LRUCache) removeCache(key interface{}) {
	entry := l.cache[key].Value.(*Entry)
	l.list.Remove(l.cache[key])
	delete(l.cache, key)
	if l.OnDroped != nil {
		l.OnDroped(key, entry.Val)
	}
}

//return number of element in cache
func (l *LRUCache) Len() int {
	if l.cache == nil {
		return 0
	}
	return l.list.Len()
}

//clear all cache
func (l *LRUCache) Clear() {
	if l.cache == nil {
		return
	}
	if l.OnDroped != nil {
		for k, v := range l.cache {
			l.OnDroped(k, v.Value.(*Entry).Val)
		}
	}
	l.cache = nil
	l.list = nil
}

func (l *LRUCache) GetAllCache() map[interface{}]*list.Element {
	return l.cache
}
