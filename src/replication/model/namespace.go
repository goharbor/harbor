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

package model

// Namespace represents the full path of resource isolation unit;
// if the namespace has hierarchical structure, e.g organization->team,
// it should be converted to organization.team
type Namespace struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

// GetStringMetadata get a string value metadata from the namespace, if not found, return the default value.
func (n *Namespace) GetStringMetadata(key string, defaultValue string) string {
	if n.Metadata == nil {
		return defaultValue
	}

	if v, ok := n.Metadata[key]; ok {
		return v.(string)
	}

	return defaultValue
}

// NamespaceQuery defines the query condition for listing namespaces
type NamespaceQuery struct {
	Name string
}
