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

package mock

import (
	"fmt"
	"reflect"

	"github.com/stretchr/testify/mock"
)

const (
	// Anything anything alias of mock.Anything
	Anything = mock.Anything
)

var (
	// AnythingOfType func alias of mock.AnythingOfType
	AnythingOfType = mock.AnythingOfType
)

// Arguments type alias of mock.Arguments
type Arguments = mock.Arguments

type mockable interface {
	On(methodName string, arguments ...interface{}) *mock.Call
}

// OnAnything mock method on obj which match any args
func OnAnything(obj interface{}, methodName string) *mock.Call {
	m, ok := obj.(mockable)
	if !ok {
		panic("obj not mockable")
	}

	v := reflect.ValueOf(obj).MethodByName(methodName)
	fnType := v.Type()

	if fnType.Kind() != reflect.Func {
		panic(fmt.Sprintf("assert: arguments: %s is not a func", v))
	}

	args := []interface{}{}
	for i := 0; i < fnType.NumIn(); i++ {
		args = append(args, mock.Anything)
	}

	return m.On(methodName, args...)
}
