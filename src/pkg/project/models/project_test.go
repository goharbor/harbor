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

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlineUpstreamRegistry(t *testing.T) {
	testCases := []struct {
		name     string
		metadata map[string]string
		expected bool
		desc     string
	}{
		{
			name:     "metadata_not_set",
			metadata: map[string]string{},
			expected: true,
			desc:     "should return true when metadata is not set (default)",
		},
		{
			name: "metadata_set_to_true",
			metadata: map[string]string{
				ProMetaUpstreamRegistryOnline: "true",
			},
			expected: true,
			desc:     "should return true when metadata is set to 'true'",
		},
		{
			name: "metadata_set_to_false",
			metadata: map[string]string{
				ProMetaUpstreamRegistryOnline: "false",
			},
			expected: false,
			desc:     "should return false when metadata is set to 'false'",
		},
		{
			name: "metadata_set_to_uppercase_true",
			metadata: map[string]string{
				ProMetaUpstreamRegistryOnline: "TRUE",
			},
			expected: true,
			desc:     "should return true when metadata is set to 'TRUE' (case-insensitive)",
		},
		{
			name: "metadata_set_to_uppercase_false",
			metadata: map[string]string{
				ProMetaUpstreamRegistryOnline: "FALSE",
			},
			expected: false,
			desc:     "should return false when metadata is set to 'FALSE' (case-insensitive)",
		},
		{
			name: "metadata_set_to_1",
			metadata: map[string]string{
				ProMetaUpstreamRegistryOnline: "1",
			},
			expected: true,
			desc:     "should return true when metadata is set to '1'",
		},
		{
			name: "metadata_set_to_0",
			metadata: map[string]string{
				ProMetaUpstreamRegistryOnline: "0",
			},
			expected: false,
			desc:     "should return false when metadata is set to '0'",
		},
		{
			name: "metadata_set_to_invalid_value",
			metadata: map[string]string{
				ProMetaUpstreamRegistryOnline: "invalid",
			},
			expected: false,
			desc:     "should return false when metadata is set to an invalid value",
		},
		{
			name: "other_metadata_present",
			metadata: map[string]string{
				ProMetaPublic: "true",
			},
			expected: true,
			desc:     "should return true when other metadata is present but upstream_registry_online is not set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			project := &Project{
				Metadata: tc.metadata,
			}
			result := project.OnlineUpstreamRegistry()
			assert.Equal(t, tc.expected, result, tc.desc)
		})
	}
}

func TestOnlineUpstreamRegistry_NilMetadata(t *testing.T) {
	// Test when Metadata is nil
	project := &Project{
		Metadata: nil,
	}
	result := project.OnlineUpstreamRegistry()
	assert.Equal(t, true, result, "should return true when metadata is nil (default)")
}

func TestOnlineUpstreamRegistry_EmptyString(t *testing.T) {
	// Test with empty string value
	project := &Project{
		Metadata: map[string]string{
			ProMetaUpstreamRegistryOnline: "",
		},
	}
	result := project.OnlineUpstreamRegistry()
	assert.Equal(t, false, result, "should return false when metadata is set to empty string")
}
