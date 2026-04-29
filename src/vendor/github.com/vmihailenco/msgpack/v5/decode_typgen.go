package msgpack

import (
	"reflect"
	"sync"
)

var cachedValues struct {
	m map[reflect.Type]chan reflect.Value
	sync.RWMutex
}

func cachedValue(t reflect.Type) reflect.Value {
	cachedValues.RLock()
	ch := cachedValues.m[t]
	cachedValues.RUnlock()
	if ch != nil {
		return <-ch
	}

	cachedValues.Lock()
	defer cachedValues.Unlock()
	if ch = cachedValues.m[t]; ch != nil {
		return <-ch
	}

	ch = make(chan reflect.Value, 256)
	go func() {
		for {
			ch <- reflect.New(t)
		}
	}()
	if cachedValues.m == nil {
		cachedValues.m = make(map[reflect.Type]chan reflect.Value, 8)
	}
	cachedValues.m[t] = ch
	return <-ch
}

func (d *Decoder) newValue(t reflect.Type) reflect.Value {
	if d.flags&usePreallocateValues == 0 {
		return reflect.New(t)
	}

	return cachedValue(t)
}
