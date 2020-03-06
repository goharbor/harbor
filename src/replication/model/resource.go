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

// the resource type
const (
	ResourceTypeArtifact ResourceType = "artifact"
	ResourceTypeImage    ResourceType = "image"
	ResourceTypeChart    ResourceType = "chart"
)

// ResourceType represents the type of the resource
type ResourceType string

// Valid indicates whether the ResourceType is a valid value
func (r ResourceType) Valid() bool {
	return len(r) > 0
}

// Resource represents the general replicating content
type Resource struct {
	Type         ResourceType           `json:"type"`
	Metadata     *ResourceMetadata      `json:"metadata"`
	Registry     *Registry              `json:"registry"`
	ExtendedInfo map[string]interface{} `json:"extended_info"`
	// Indicate if the resource is a deleted resource
	Deleted bool `json:"deleted"`
	// indicate whether the resource can be overridden
	Override bool `json:"override"`
}

// ResourceMetadata of resource
type ResourceMetadata struct {
	Repository *Repository `json:"repository"`
	Artifacts  []*Artifact `json:"artifacts"`
	Vtags      []string    `json:"v_tags"` // deprecated, use Artifacts instead
}

// Repository info of the resource
type Repository struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Artifact is the individual unit that can be replicated
type Artifact struct {
	Type   string   `json:"type"`
	Digest string   `json:"digest"`
	Labels []string `json:"labels"`
	Tags   []string `json:"tags"`
}
