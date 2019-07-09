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

package action

import (
	"github.com/pkg/errors"
	"sync"
)

// index for keeping the mapping action and its performer
var index sync.Map

// Register the performer with the corresponding action
func Register(action string, factory PerformerFactory) {
	if len(action) == 0 || factory == nil {
		// do nothing
		return
	}

	index.Store(action, factory)
}

// Get performer with the provided action
func Get(action string) (Performer, error) {
	if len(action) == 0 {
		return nil, errors.New("empty action")
	}

	v, ok := index.Load(action)
	if !ok {
		return nil, errors.Errorf("action %s is not registered", action)
	}

	factory, ok := v.(PerformerFactory)
	if !ok {
		return nil, errors.Errorf("invalid action performer registered for action %s", action)
	}

	return factory(), nil
}
