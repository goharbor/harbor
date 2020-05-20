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
	"errors"
	"fmt"
	"path"
	"strings"
)

// Resource the type of resource
type Resource string

func (res Resource) String() string {
	return string(res)
}

// RelativeTo returns relative resource to other resource
func (res Resource) RelativeTo(other Resource) (Resource, error) {
	prefix := other.String()
	str := res.String()

	if !strings.HasPrefix(str, prefix) {
		return Resource(""), errors.New("value error")
	}

	relative := strings.TrimPrefix(strings.TrimPrefix(str, prefix), "/")
	if relative == "" {
		relative = "."
	}

	return Resource(relative), nil
}

// Subresource returns subresource
func (res Resource) Subresource(resources ...Resource) Resource {
	elements := []string{res.String()}

	for _, resource := range resources {
		elements = append(elements, resource.String())
	}

	return Resource(path.Join(elements...))
}

// GetNamespace returns namespace from resource
func (res Resource) GetNamespace() (Namespace, error) {
	return nil, fmt.Errorf("no namespace found for %s", res)
}
