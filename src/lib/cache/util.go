// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"errors"
	"net/url"
	"reflect"
	"sync"
)

// TODO: use the URL.Redacted when golang upgrade to 1.15
func redacted(u *url.URL) string {
	if u == nil {
		return ""
	}

	ru := *u
	if _, has := ru.User.Password(); has {
		ru.User = url.UserPassword(ru.User.Username(), "xxxxx")
	}
	return ru.String()
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func simpleCopy(dst interface{}, src interface{}) error {
	from := indirect(reflect.ValueOf(src))
	to := indirect(reflect.ValueOf(dst))
	if !to.CanAddr() {
		return errors.New("dst value is unaddressable")
	}

	if !from.Type().ConvertibleTo(to.Type()) {
		return errors.New("src value is not convertible to the dst value")
	}

	to.Set(from.Convert(to.Type()))

	return nil
}

type keyMutex struct {
	m *sync.Map
}

func (km keyMutex) Lock(key interface{}) {
	m := sync.Mutex{}
	act, _ := km.m.LoadOrStore(key, &m)

	mm := act.(*sync.Mutex)
	mm.Lock()
	if mm != &m {
		mm.Unlock()
		km.Lock(key)
		return
	}

	return
}

func (km keyMutex) Unlock(key interface{}) {
	act, exist := km.m.Load(key)
	if !exist {
		panic("unlock of unlocked mutex")
	}
	m := act.(*sync.Mutex)
	km.m.Delete(key)
	m.Unlock()
}
