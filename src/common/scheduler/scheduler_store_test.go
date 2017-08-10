package scheduler

import (
	"testing"
)

func TestPut(t *testing.T) {
	store := NewDefaultStore()
	if store == nil {
		t.Fatal("Failed to creat store instance")
	}

	store.Put("testing", NewWatcher(nil, nil, nil))
	if store.Size() != 1 {
		t.Fail()
	}
}

func TestGet(t *testing.T) {
	store := NewDefaultStore()
	if store == nil {
		t.Fatal("Failed to creat store instance")
	}
	store.Put("testing", NewWatcher(nil, nil, nil))
	w := store.Get("testing")
	if w == nil {
		t.Fail()
	}
}

func TestRemove(t *testing.T) {
	store := NewDefaultStore()
	if store == nil {
		t.Fatal("Failed to creat store instance")
	}
	store.Put("testing", NewWatcher(nil, nil, nil))
	if !store.Exists("testing") {
		t.Fail()
	}
	w := store.Remove("testing")
	if w == nil {
		t.Fail()
	}
}

func TestExisting(t *testing.T) {
	store := NewDefaultStore()
	if store == nil {
		t.Fatal("Failed to creat store instance")
	}
	store.Put("testing", NewWatcher(nil, nil, nil))
	if !store.Exists("testing") {
		t.Fail()
	}
	if store.Exists("fake_key") {
		t.Fail()
	}
}

func TestGetAll(t *testing.T) {
	store := NewDefaultStore()
	if store == nil {
		t.Fatal("Failed to creat store instance")
	}
	store.Put("testing", NewWatcher(nil, nil, nil))
	store.Put("testing2", NewWatcher(nil, nil, nil))
	list := store.GetAll()
	if list == nil || len(list) != 2 {
		t.Fail()
	}
}
