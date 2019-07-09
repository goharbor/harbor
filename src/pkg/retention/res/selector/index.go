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

package selector

import (
	"fmt"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/pkg/errors"
	"sync"
)

// index for keeping the mapping between selector meta and its implementation
var index sync.Map

// Register the selector with the corresponding selector kind and decoration
func Register(kind, decoration string, factory res.SelectorFactory) {
	id := fmt.Sprintf("%s:%s", kind, decoration)
	if len(id) == 0 || factory == nil {
		// do nothing
		return
	}

	index.Store(id, factory)
}

// Get selector with the provided kind and decoration
func Get(kind, decoration string, pattern interface{}) (res.Selector, error) {
	if len(templateID) == 0 {
		return nil, errors.New("empty rule template ID")
	}

	v, ok := index.Load(templateID)
	if !ok {
		return nil, errors.Errorf("rule evaluator %s is not registered", templateID)
	}

	factory, ok := v.(RuleFactory)
	if !ok {
		return nil, errors.Errorf("invalid rule evaluator registered for %s", templateID)
	}

	return factory(parameters), nil
}
