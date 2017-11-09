package trigger

import "testing"
import "time"

func TestCache(t *testing.T) {
	cache := NewCache(10)
	trigger := NewImmediateTrigger(ImmediateParam{})

	cache.Put(1, trigger)
	if cache.Size() != 1 {
		t.Fatalf("Invalid size, expect 1 but got %d", cache.Size())
	}

	tr := cache.Get(1)
	if tr == nil {
		t.Fatal("Should not get nil item")
	}

	tri := cache.Remove(1)
	if tri == nil || cache.Size() > 0 {
		t.Fatal("Failed to remove")
	}
}

func TestCacheChange(t *testing.T) {
	cache := NewCache(2)
	trigger1 := NewImmediateTrigger(ImmediateParam{})
	trigger2 := NewImmediateTrigger(ImmediateParam{})
	cache.Put(1, trigger1)
	cache.Put(2, trigger2)

	if cache.Size() != 2 {
		t.Fatalf("Invalid size, expect 2 but got %d", cache.Size())
	}

	if tr := cache.Get(2); tr == nil {
		t.Fatal("Should not get nil item")
	}

	time.Sleep(100 * time.Microsecond)

	trigger3 := NewImmediateTrigger(ImmediateParam{})
	cache.Put(3, trigger3)
	if cache.Size() != 2 {
		t.Fatalf("Invalid size, expect 2 but got %d", cache.Size())
	}

	if tr := cache.Get(1); tr != nil {
		t.Fatal("item1 should not exist")
	}

}
