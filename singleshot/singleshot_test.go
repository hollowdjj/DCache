package singleshot

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	var count int32
	ch := make(chan string)
	fn := func() (interface{}, error) {
		atomic.AddInt32(&count, 1)
		return <-ch, nil
	}

	shots := Shots{}
	wg := sync.WaitGroup{}
	n := 100
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := shots.Do("key", fn)
			if err != nil {
				t.Errorf("got error : %v", err)
			}
			if v.(string) != "done" {
				t.Errorf("got %v but want %v", v, "done")
			}
		}()
	}
	time.Sleep(time.Millisecond * 100)
	ch <- "done"
	if got := atomic.LoadInt32(&count); got != 1 {
		t.Errorf("got %v but want %v", got, 1)
	}
}
