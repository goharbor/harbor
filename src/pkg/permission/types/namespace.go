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

package types

import (
	"sync"
)

var (
	parsesMu sync.RWMutex
	parses   = map[string]NamespaceParse{}
)

// NamespaceParse parse namespace from the resource
type NamespaceParse func(Resource) (Namespace, bool)

// Namespace the namespace interface
type Namespace interface {
	// Kind returns the kind of namespace
	Kind() string
	// Resource returns new resource for subresources with the namespace
	Resource(subresources ...Resource) Resource
	// Identity returns identity attached with namespace
	Identity() interface{}
	// GetPolicies returns all policies of the namespace
	GetPolicies() []*Policy
}

// RegistryNamespaceParse ...
func RegistryNamespaceParse(name string, parse NamespaceParse) {
	parsesMu.Lock()
	defer parsesMu.Unlock()
	if parse == nil {
		panic("permission: Register namespace parse is nil")
	}
	if _, dup := parses[name]; dup {
		panic("permission: Register called twice for namespace parse " + name)
	}

	parses[name] = parse
}

// NamespaceFromResource returns namespace from resource
func NamespaceFromResource(resource Resource) (Namespace, bool) {
	parsesMu.RLock()
	defer parsesMu.RUnlock()

	for _, parse := range parses {
		if ns, ok := parse(resource); ok {
			return ns, true
		}
	}

	return nil, false
}

// ResourceAllowedInNamespace returns true when resource's namespace equal the ns
func ResourceAllowedInNamespace(resource Resource, ns Namespace) bool {
	n, ok := NamespaceFromResource(resource)
	if ok {
		return n.Kind() == ns.Kind() && n.Identity() == ns.Identity()
	}

	return false
}
