package lru

import "testing"

func TestLRUCache(t *testing.T) {
	//测试Add
	c1 := New()
	var c2 LRUCache
	for i := 0; i < 5; i++ {
		c1.Add(i, i)
		c2.Add(i, i)
	}
	if c1.Len() != 3 || c2.Len() != 5 {
		t.Error("Bugs in function Add")
	}
	if _, hit := c1.Get(0); hit {
		t.Error("Bugs in function Add")
	}
	if _, hit := c1.Get(1); hit {
		t.Error("Bugs in function Add")
	}
	c1.Add(4, 5)
	e := c1.list.Front().Value.(*entry)
	if e.key != 4 || e.val != 5 {
		t.Error("Bugs in function Add")
	}
}
