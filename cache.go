package cache

import (
	"sync"
	"time"

	"github.com/hollowdjj/course-selecting-sys/cache/lru"
)

var (
	delTTL   = 100 //unit: ms
	delCount = 20
)

//encapsulation of LRU cache, concurrency safe
type cache struct {
	rw           sync.RWMutex
	lru          *lru.LRUCache //LRU cache
	nbytes       int64         //memory usage of cache
	ngets, nhits int64
}

//create a new concurrency safe cache
func NewCache() *cache {
	res := &cache{lru: lru.New()}
	res.timingDel()
	return res
}

//add cache, concurrency safe
func (c *cache) add(key string, val Value, ttl time.Duration) {
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
	c.lru.Add(key, val, ttl)
	c.nbytes += int64(len(key)) + int64(val.Len())
}

//get cache, concurrency safe
func (c *cache) get(key string) (value Value, ok bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if c.lru == nil {
		return
	}
	c.ngets++
	entry, hit := c.lru.Get(key)
	if !hit {
		return
	}
	c.nhits++
	//delete if expired
	if time.Now().After(entry.ExpireAt) {
		c.lru.Del(entry.Key)
		return Value{}, false
	}

	return entry.Val.(Value), true
}

//del cache, concurrency safe
func (c *cache) del(key string) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if c.lru == nil {
		return
	}
	c.lru.Del(key)
}

//remove least recently used cache
func (c *cache) removeLeastUsed() int64 {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.lru.RemoveLeastUsed()
	return c.nbytes
}

//every [delTTL] millisecond, randomly select [delCount] cache and delect it when expire
func (c *cache) timingDel() {
	go func() {
		ticker := time.NewTicker(time.Duration(delTTL) * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				c.rw.Lock()
				count := 0
				for k, v := range c.lru.GetAllCache() {
					//delete if expired
					nodeEntry := v.Value.(*lru.Entry)
					if time.Now().After(nodeEntry.ExpireAt) {
						c.lru.Del(k)
					}
					count++
					if count >= delCount {
						break
					}
				}
				c.rw.Unlock()
			}
		}
	}()
}
