package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	//测试Add
	c1 := NewCache()
	c1.add("1", Value{}, 2*time.Second)
	fmt.Println(time.Now())
	time.Sleep(1 * time.Second)
	if _, ok := c1.get("1"); ok {
		t.Errorf("bug")
	}
}
