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

package art

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"

	"github.com/pkg/errors"
)

const (
	// Image kind
	Image = "image"
	// Chart kind
	Chart = "chart"
)

// Repository of candidate
type Repository struct {
	// Namespace(project) ID
	NamespaceID int64
	// Namespace
	Namespace string `json:"namespace"`
	// Repository name
	Name string `json:"name"`
	// So far we need the kind of repository and retrieve candidates with different APIs
	// TODO: REMOVE IT IN THE FUTURE IF WE SUPPORT UNIFIED ARTIFACT MODEL
	Kind string `json:"kind"`
}

// ToJSON marshals repository to JSON string
func (r *Repository) ToJSON() (string, error) {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return "", errors.Wrap(err, "marshal reporitory")
	}

	return string(jsonData), nil
}

// FromJSON constructs the repository from json data
func (r *Repository) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to construct repository")
	}

	return json.Unmarshal([]byte(jsonData), r)
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
	// Digest
	Digest string
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
	if c.Digest == "" {
		log.Errorf("Lack Digest of Candidate for %s/%s:%s", c.Namespace, c.Repository, c.Tag)
	}
	raw := fmt.Sprintf("%s:%s/%s:%s", c.Kind, c.Namespace, c.Repository, c.Digest)

	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// NameHash based on the candidate info for differentiation
func (c *Candidate) NameHash() string {
	raw := fmt.Sprintf("%s:%s/%s:%s", c.Kind, c.Namespace, c.Repository, c.Tag)

	return base64.StdEncoding.EncodeToString([]byte(raw))
}
