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

package proxy

import "sync"

type inflightRequest struct {
	mu     sync.Mutex
	reqMap map[string]chan struct{}
}

var inflightChecker = &inflightRequest{
	reqMap: make(map[string]chan struct{}),
}

// addRequest registers the artifact as in flight and returns true if the
// caller now owns the request. If it is already in flight, returns false and
// a channel that is closed once the owner removes the request.
func (in *inflightRequest) addRequest(artifact string) (suc bool, done <-chan struct{}) {
	in.mu.Lock()
	defer in.mu.Unlock()
	if ch, ok := in.reqMap[artifact]; ok {
		return false, ch
	}
	ch := make(chan struct{})
	in.reqMap[artifact] = ch
	return true, ch
}

func (in *inflightRequest) removeRequest(artifact string) {
	in.mu.Lock()
	defer in.mu.Unlock()
	if ch, ok := in.reqMap[artifact]; ok {
		close(ch)
		delete(in.reqMap, artifact)
	}
}
