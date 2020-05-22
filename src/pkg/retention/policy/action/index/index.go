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

package index

import (
	"sync"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
)

// index for keeping the mapping action and its performer
var index sync.Map

func init() {
	// Register retain action
	Register(action.Retain, action.NewRetainAction)
}

// Register the performer with the corresponding action
func Register(action string, factory action.PerformerFactory) {
	if len(action) == 0 || factory == nil {
		// do nothing
		return
	}

	index.Store(action, factory)
}

// Get performer with the provided action
func Get(act string, params interface{}, isDryRun bool) (action.Performer, error) {
	if len(act) == 0 {
		return nil, errors.New("empty action")
	}

	v, ok := index.Load(act)
	if !ok {
		return nil, errors.Errorf("action %s is not registered", act)
	}

	factory, ok := v.(action.PerformerFactory)
	if !ok {
		return nil, errors.Errorf("invalid action performer registered for action %s", act)
	}

	return factory(params, isDryRun), nil
}
