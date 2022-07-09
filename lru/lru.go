package lru

import (
	"container/list"
)

type LRUCache struct {
	//最大缓存元素数量，0表示无上限(但仍然会受到系统内存限制)
	MaxEntries int

	//双向链表+哈希表实现LRU缓存
	list  *list.List
	cache map[interface{}]*list.Element

	//元素删除时的回调函数
	OnDroped func(key interface{}, val interface{})
}

//链表节点存储值
type entry struct {
	key interface{}
	val interface{}
}

//生成一个LRU缓存。
//若MaxEntries为0，表示缓存元素无上限(但仍会受到系统内存限制)
func New(MaxEntries int) *LRUCache {
	return &LRUCache{
		MaxEntries: MaxEntries,
		list:       list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

//向缓存中添加新元素
func (l *LRUCache) Add(key interface{}, val interface{}) {
	if l.cache == nil {
		l.cache = make(map[interface{}]*list.Element)
		l.list = list.New()
	}

	//若key存在，将节点移至链表头并更新值
	node, ok := l.cache[key]
	if ok {
		l.list.MoveToFront(node)
		node.Value.(*entry).val = val
		return
	}

	//插入新元素
	newNode := l.list.PushFront(&entry{key, val})
	l.cache[key] = newNode

	//缓存淘汰
	if l.MaxEntries > 0 && l.list.Len() > l.MaxEntries {
		l.RemoveLeastUsed()
	}
}

//根据key查找缓存
func (l *LRUCache) Get(key interface{}) (val interface{}, ok bool) {
	if l.cache == nil {
		return
	}
	if node, hit := l.cache[key]; hit {
		l.list.MoveToFront(node)
		return node.Value.(*entry).val, true
	}
	return
}

//根据key删除缓存
func (l *LRUCache) Remove(key interface{}) {
	if l.cache == nil {
		return
	}
	if _, hit := l.cache[key]; hit {
		l.removeCache(key)
	}
}

//删除最近最少使用的元素
func (l *LRUCache) RemoveLeastUsed() {
	if l.cache == nil {
		return
	}
	target := l.list.Back()
	if target != nil {
		l.list.Remove(target)
		key := target.Value.(*entry).key
		delete(l.cache, key)
	}
}

//删除缓存
func (l *LRUCache) removeCache(key interface{}) {
	entry := l.cache[key].Value.(*entry)
	l.list.Remove(l.cache[key])
	delete(l.cache, key)
	if l.OnDroped != nil {
		l.OnDroped(key, entry.val)
	}
}

//缓存中的元素数量
func (l *LRUCache) Len() int {
	if l.cache == nil {
		return 0
	}
	return l.list.Len()
}

//清空缓存
func (l *LRUCache) Clear() {
	if l.cache == nil {
		return
	}
	if l.OnDroped != nil {
		for k, v := range l.cache {
			l.OnDroped(k, v.Value.(*entry).val)
		}
	}
	l.cache = nil
	l.list = nil
}
