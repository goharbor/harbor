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

package project

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/goharbor/harbor/src/pkg/permission/types"
)

const (
	// NamespaceKind kind for project projectNamespace
	NamespaceKind = "project"
)

var (
	projectNamespaceRe = regexp.MustCompile("^/project/([^/]*)/?")
)

type projectNamespace struct {
	projectID int64
}

func (ns *projectNamespace) Kind() string {
	return NamespaceKind
}

func (ns *projectNamespace) Resource(subresources ...types.Resource) types.Resource {
	return types.Resource(fmt.Sprintf("/project/%d", ns.projectID)).Subresource(subresources...)
}

func (ns *projectNamespace) Identity() interface{} {
	return ns.projectID
}

func (ns *projectNamespace) GetPolicies() []*types.Policy {
	return GetPoliciesOfProject(ns.projectID)
}

// NewNamespace returns projectNamespace for project
func NewNamespace(projectID int64) types.Namespace {
	return &projectNamespace{projectID: projectID}
}

// NamespaceParse ...
func NamespaceParse(resource types.Resource) (types.Namespace, bool) {
	matches := projectNamespaceRe.FindStringSubmatch(resource.String())

	if len(matches) <= 1 {
		return nil, false
	}

	projectID, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return nil, false
	}

	return NewNamespace(projectID), true
}

func init() {
	types.RegistryNamespaceParse(NamespaceKind, NamespaceParse)
}
