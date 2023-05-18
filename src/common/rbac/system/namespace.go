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

package system

import (
	"strings"

	"github.com/goharbor/harbor/src/pkg/permission/types"
)

const (
	// NamespaceKind kind for system namespace
	NamespaceKind = "system"
	// NamespacePrefix for system namespace
	NamespacePrefix = "/system"
)

type systemNamespace struct {
}

func (ns *systemNamespace) Kind() string {
	return NamespaceKind
}

func (ns *systemNamespace) Resource(subresources ...types.Resource) types.Resource {
	return types.Resource("/system/").Subresource(subresources...)
}

func (ns *systemNamespace) Identity() interface{} {
	return nil
}

func (ns *systemNamespace) GetPolicies() []*types.Policy {
	return policies
}

// NewNamespace returns namespace for project
func NewNamespace() types.Namespace {
	return &systemNamespace{}
}

// NamespaceParse ...
func NamespaceParse(resource types.Resource) (types.Namespace, bool) {
	if strings.HasPrefix(resource.String(), NamespacePrefix) {
		return NewNamespace(), true
	}
	return nil, false
}

func init() {
	types.RegistryNamespaceParse(NamespaceKind, NamespaceParse)
}
