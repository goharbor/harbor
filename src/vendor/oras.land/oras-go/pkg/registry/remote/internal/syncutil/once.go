/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package syncutil

import "context"

// Once is an object that will perform exactly one action.
// Unlike sync.Once, this Once allowes the action to have return values.
type Once struct {
	result interface{}
	err    error
	status chan bool
}

// NewOnce creates a new Once instance.
func NewOnce() *Once {
	status := make(chan bool, 1)
	status <- true
	return &Once{
		status: status,
	}
}

// Do calls the function f if and only if Do is being called first time or all
// previous function calls are cancelled, deadline exceeded, or panicking.
// When `once.Do(ctx, f)` is called multiple times, the return value of the
// first call of the function f is stored, and is directly returned for other
// calls.
// Besides the return value of the function f, including the error, Do returns
// true if the function f passed is called first and is not cancelled, deadline
// exceeded, or panicking. Otherwise, returns false.
func (o *Once) Do(ctx context.Context, f func() (interface{}, error)) (bool, interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			o.status <- true
			panic(r)
		}
	}()
	for {
		select {
		case inProgress := <-o.status:
			if !inProgress {
				return false, o.result, o.err
			}
			result, err := f()
			if err == context.Canceled || err == context.DeadlineExceeded {
				o.status <- true
				return false, nil, err
			}
			o.result, o.err = result, err
			close(o.status)
			return true, result, err
		case <-ctx.Done():
			return false, nil, ctx.Err()
		}
	}
}
