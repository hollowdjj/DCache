package singleshot

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup
	val interface{} //函数返回值，一个空interface以及一个error
	err error
}

//用于避免缓存击穿(某一热点key过期，瞬间大量请求打到数据库上)
type Shots struct {
	mu  sync.Mutex
	dic map[string]*call
}

func (s *Shots) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	s.mu.Lock()
	//延迟初始化
	if s.dic == nil {
		s.dic = make(map[string]*call)
	}

	//针对key，有一个请求在进行中，调用Wait阻塞
	if c, ok := s.dic[key]; ok {
		s.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	//创建call实例
	c := &call{}
	c.wg.Add(1)
	s.dic[key] = c
	s.mu.Unlock()

	//调用fn函数
	c.val, c.err = fn()
	c.wg.Done()

	//删除key-call
	s.mu.Lock()
	delete(s.dic, key)
	s.mu.Unlock()

	return c.val, c.err
}
