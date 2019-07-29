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

// Package lwp = lightweight policy
package lwp

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
)

// Metadata contains partial metadata of policy
// It's a lightweight version of policy.Metadata
type Metadata struct {
	// Algorithm applied to the rules
	// "OR" / "AND"
	Algorithm string `json:"algorithm"`

	// Rule collection
	Rules []*rule.Metadata `json:"rules"`
}

// ToJSON marshals metadata to JSON string
func (m *Metadata) ToJSON() (string, error) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return "", errors.Wrap(err, "marshal reporitory")
	}

	return string(jsonData), nil
}

// FromJSON constructs the metadata from json data
func (m *Metadata) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to construct repository")
	}

	return json.Unmarshal([]byte(jsonData), m)
}
