package lru

import (
	"fmt"
	"testing"
	"time"
)

func TestLRUCache(t *testing.T) {
	//测试Add
	c1 := New()
	c1.Add(1, 1, 2*time.Second)
	fmt.Println(time.Now())
	time.Sleep(3 * time.Second)
	if node, ok := c1.Get(1); ok {
		t.Errorf("get %s", node.ExpireAt)
	}
}
