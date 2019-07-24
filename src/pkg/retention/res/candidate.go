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

package res

import (
	"encoding/base64"
	"fmt"
)

const (
	// Image kind
	Image = "image"
	// Chart kind
	Chart = "chart"
)

// Repository of candidate
type Repository struct {
	// Namespace
	Namespace string
	// Repository name
	Name string
	// So far we need the kind of repository and retrieve candidates with different APIs
	// TODO: REMOVE IT IN THE FUTURE IF WE SUPPORT UNIFIED ARTIFACT MODEL
	Kind string
}

// Candidate for retention processor to match
type Candidate struct {
	// Namespace(project) ID
	NamespaceID int64
	// Namespace
	Namespace string
	// Repository name
	Repository string
	// Kind of the candidate
	// "image" or "chart"
	Kind string
	// Tag info
	Tag string
	// Pushed time in seconds
	PushedTime int64
	// Pulled time in seconds
	PulledTime int64
	// Created time in seconds
	CreationTime int64
	// Labels attached with the candidate
	Labels []string
}

// Hash code based on the candidate info for differentiation
func (c *Candidate) Hash() string {
	raw := fmt.Sprintf("%s:%s/%s:%s", c.Kind, c.Namespace, c.Repository, c.Tag)

	return base64.StdEncoding.EncodeToString([]byte(raw))
}
