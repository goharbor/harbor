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

package rbac

import (
	"fmt"
)

// Namespace the namespace interface
type Namespace interface {
	// Kind returns the kind of namespace
	Kind() string
	// Resource returns new resource for subresources with the namespace
	Resource(subresources ...Resource) Resource
	// Identity returns identity attached with namespace
	Identity() interface{}
	// IsPublic returns true if namespace is public
	IsPublic() bool
}

type projectNamespace struct {
	projectID int64
	isPublic  bool
}

func (ns *projectNamespace) Kind() string {
	return "project"
}

func (ns *projectNamespace) Resource(subresources ...Resource) Resource {
	return Resource(fmt.Sprintf("/project/%d", ns.projectID)).Subresource(subresources...)
}

func (ns *projectNamespace) Identity() interface{} {
	return ns.projectID
}

func (ns *projectNamespace) IsPublic() bool {
	return ns.isPublic
}

// NewProjectNamespace returns namespace for project
func NewProjectNamespace(projectID int64, isPublic ...bool) Namespace {
	isPublicNamespace := false
	if len(isPublic) > 0 {
		isPublicNamespace = isPublic[0]
	}
	return &projectNamespace{projectID: projectID, isPublic: isPublicNamespace}
}
