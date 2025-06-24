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

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResource_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resource Resource
		wantJSON string
	}{
		{
			name: "scan_overview with nil value",
			resource: Resource{
				Digest:      "sha256:123456",
				Tag:         "latest",
				ResourceURL: "http://example.com/resource",
			},
			wantJSON: `{
  "digest": "sha256:123456",
  "tag": "latest",
  "resource_url": "http://example.com/resource",
  "scan_overview": null
}`,
		},
		{
			name: "scan_overview with empty map",
			resource: Resource{
				Digest:       "sha256:123456",
				Tag:          "latest",
				ResourceURL:  "http://example.com/resource",
				ScanOverview: map[string]any{},
			},
			wantJSON: `{
  "digest": "sha256:123456",
  "tag": "latest",
  "resource_url": "http://example.com/resource",
  "scan_overview": {}
}`,
		},
		{
			name: "scan_overview with data",
			resource: Resource{
				Digest:      "sha256:123456",
				Tag:         "latest",
				ResourceURL: "http://example.com/resource",
				ScanOverview: map[string]any{
					"application/vnd.security.vulnerability.report; version=1.1": map[string]any{
						"severity":   "High",
						"scan_status": "Success",
						"total_count": 5,
					},
				},
			},
			wantJSON: `{
  "digest": "sha256:123456",
  "tag": "latest",
  "resource_url": "http://example.com/resource",
  "scan_overview": {
    "application/vnd.security.vulnerability.report; version=1.1": {
      "scan_status": "Success",
      "severity": "High",
      "total_count": 5
    }
  }
}`,
		},
		{
			name: "sbom_overview with nil value should be omitted",
			resource: Resource{
				Digest:       "sha256:123456",
				Tag:          "latest",
				ResourceURL:  "http://example.com/resource",
				ScanOverview: map[string]any{},
			},
			wantJSON: `{
  "digest": "sha256:123456",
  "tag": "latest",
  "resource_url": "http://example.com/resource",
  "scan_overview": {}
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.MarshalIndent(tt.resource, "", "  ")
			require.NoError(t, err)

			// Parse both expected and actual JSON to compare them semantically
			var expectedJSON, actualJSON interface{}
			err = json.Unmarshal([]byte(tt.wantJSON), &expectedJSON)
			require.NoError(t, err)
			err = json.Unmarshal(jsonBytes, &actualJSON)
			require.NoError(t, err)

			assert.Equal(t, expectedJSON, actualJSON, "JSON output doesn't match expected")

			// Verify scan_overview is always present in the JSON
			var jsonMap map[string]interface{}
			err = json.Unmarshal(jsonBytes, &jsonMap)
			require.NoError(t, err)
			_, exists := jsonMap["scan_overview"]
			assert.True(t, exists, "scan_overview field should always be present in JSON")
		})
	}
}

func TestResource_ScanOverviewAlwaysPresent(t *testing.T) {
	// This test specifically verifies that scan_overview is never omitted
	// even when it's nil, which is the fix for the Zulip webhook issue
	resource := Resource{
		Digest:      "sha256:test",
		Tag:         "v1.0",
		ResourceURL: "http://harbor.example.com/project/repo:v1.0",
		// ScanOverview is intentionally left as nil (zero value)
	}

	jsonBytes, err := json.Marshal(resource)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	// The key assertion: scan_overview must be present even when nil
	scanOverview, exists := result["scan_overview"]
	assert.True(t, exists, "scan_overview field must always be present in JSON output")
	assert.Nil(t, scanOverview, "scan_overview should be null when not initialized")
}