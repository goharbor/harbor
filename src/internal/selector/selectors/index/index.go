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
	"github.com/goharbor/harbor/src/internal/selector"
	"sync"

	"github.com/goharbor/harbor/src/internal/selector/selectors/doublestar"
	"github.com/pkg/errors"
)

func init() {
	// Register doublestar selector
	Register(doublestar.Kind, []string{
		doublestar.Matches,
		doublestar.Excludes,
		doublestar.RepoMatches,
		doublestar.RepoExcludes,
		doublestar.NSMatches,
		doublestar.NSExcludes,
	}, doublestar.New)

	// Register label selector
	// Register(label.Kind, []string{label.With, label.Without}, label.New)
}

// index for keeping the mapping between selector meta and its implementation
var index sync.Map

// IndexedMeta describes the indexed selector
type IndexedMeta struct {
	Kind        string   `json:"kind"`
	Decorations []string `json:"decorations"`
}

// indexedItem defined item kept in the index
type indexedItem struct {
	Meta    *IndexedMeta
	Factory selector.Factory
}

// Register the selector with the corresponding selector kind and decoration
func Register(kind string, decorations []string, factory selector.Factory) {
	if len(kind) == 0 || factory == nil {
		// do nothing
		return
	}

	index.Store(kind, &indexedItem{
		Meta: &IndexedMeta{
			Kind:        kind,
			Decorations: decorations,
		},
		Factory: factory,
	})
}

// Get selector with the provided kind and decoration
func Get(kind, decoration, pattern, extras string) (selector.Selector, error) {
	if len(kind) == 0 || len(decoration) == 0 {
		return nil, errors.New("empty selector kind or decoration")
	}

	v, ok := index.Load(kind)
	if !ok {
		return nil, errors.Errorf("selector %s is not registered", kind)
	}

	item := v.(*indexedItem)
	for _, dec := range item.Meta.Decorations {
		if dec == decoration {
			factory := item.Factory
			return factory(decoration, pattern, extras), nil
		}
	}

	return nil, errors.Errorf("decoration %s of selector %s is not supported", decoration, kind)
}

// Index returns all the declarative selectors
func Index() []*IndexedMeta {
	all := make([]*IndexedMeta, 0)

	index.Range(func(k, v interface{}) bool {
		if item, ok := v.(*indexedItem); ok {
			all = append(all, item.Meta)
			return true
		}

		return false
	})

	return all
}
