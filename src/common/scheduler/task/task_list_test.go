package task

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestTaskList(t *testing.T) {
	ds := NewDefaultStore()
	if ds.tasks == nil {
		t.Fatal("Failed to create store")
	}

	go func() {
		var count int32
		for {
			ds.AddTasks(NewScanAllTask())
			atomic.AddInt32(&count, 1)
			time.Sleep(100 * time.Millisecond)
			if atomic.LoadInt32(&count) > 9 {
				return
			}
		}
	}()
	go func() {
		var count int32
		for {
			ds.GetTasks()
			atomic.AddInt32(&count, 1)
			time.Sleep(100 * time.Millisecond)
			if atomic.LoadInt32(&count) > 8 {
				return
			}
		}
	}()

	<-time.After(2 * time.Second)

	var taskCount int32
	atomic.StoreInt32(&taskCount, (int32)(len(ds.GetTasks())))

	if atomic.LoadInt32(&taskCount) != 10 {
		t.Fatalf("Expect %d tasks but got %d", 10, atomic.LoadInt32(&taskCount))
	}
}
