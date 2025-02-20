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

package event

import (
	"encoding/json"
	"slices"

	"github.com/goharbor/harbor/src/lib/log"
)

// Redact replaces sensitive attributes in the JSON payload with "***"
func Redact(payload string, sensitiveAttributes []string) string {
	if len(payload) == 0 {
		return ""
	}
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &jsonData); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
		return ""
	}
	replacePassword(jsonData, sensitiveAttributes)
	// Convert the modified map back to JSON
	modifiedJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		log.Fatalf("Error converting to JSON: %v", err)
		return ""
	}
	return string(modifiedJSON)
}

// replacePassword recursively replaces attribute in maskAttributes's value with "***"
func replacePassword(data map[string]interface{}, maskAttributes []string) {
	for key, value := range data {
		if slices.Contains(maskAttributes, key) {
			data[key] = "***"
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			replacePassword(nestedMap, maskAttributes)
		} else if nestedArray, ok := value.([]interface{}); ok {
			for _, item := range nestedArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					replacePassword(itemMap, maskAttributes)
				}
			}
		}
	}
}
