package dcache

import (
	"github/hollowdjj/DCache/lru"
	"sync"
)

//对LRU缓存进行封装，使其并发安全
type cache struct {
	rw           sync.RWMutex
	lru          *lru.LRUCache
	nbytes       int64
	ngets, nhits int64
}

//添加缓存
func (c *cache) add(key string, val Value) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if c.lru == nil {
		c.lru = &lru.LRUCache{
			OnDroped: func(key interface{}, value interface{}) {
				val := value.(Value)
				c.nbytes -= int64(len(key.(string))) + int64(val.Len())
			},
		}
	}
	c.lru.Add(key, val)
	c.nbytes += int64(len(key)) + int64(val.Len())
}

//获取缓存
func (c *cache) get(key string) (value Value, ok bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	c.ngets++
	if c.lru == nil {
		return
	}
	val, ok := c.lru.Get(key)
	if !ok {
		return
	}
	c.nhits++

	return val.(Value), true
}
