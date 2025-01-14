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

package provider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstance_FromJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "Empty JSON",
			json:    "",
			wantErr: true,
		},
		{
			name:    "Invalid JSON",
			json:    "{invalid}",
			wantErr: true,
		},
		{
			name: "Valid JSON",
			json: `{
				"id": 1,
				"name": "test-instance",
				"description": "test description",
				"vendor": "test-vendor",
				"endpoint": "http://test-endpoint",
				"auth_mode": "basic",
				"auth_info": {"username": "test", "password": "test123"},
				"enabled": true,
				"default": false,
				"insecure": false,
				"setup_timestamp": 1234567890
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ins := &Instance{}
			err := ins.FromJSON(tt.json)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), ins.ID)
				assert.Equal(t, "test-instance", ins.Name)
			}
		})
	}
}

func TestInstance_ToJSON(t *testing.T) {
	ins := &Instance{
		ID:          1,
		Name:        "test-instance",
		Description: "test description",
		Vendor:      "test-vendor",
		Endpoint:    "http://test-endpoint",
		AuthMode:    "basic",
		AuthInfo:    map[string]string{"username": "test", "password": "test123"},
		Enabled:     true,
		Default:     false,
		Insecure:    false,
	}

	jsonStr, err := ins.ToJSON()
	assert.NoError(t, err)

	// Verify the JSON can be decoded back
	var decoded Instance
	err = json.Unmarshal([]byte(jsonStr), &decoded)
	assert.NoError(t, err)
	assert.Equal(t, ins.ID, decoded.ID)
	assert.Equal(t, ins.Name, decoded.Name)
	assert.Equal(t, ins.AuthInfo, decoded.AuthInfo)
}

func TestInstance_Decode(t *testing.T) {
	tests := []struct {
		name     string
		authData string
		wantErr  bool
	}{
		{
			name:     "Empty auth data",
			authData: "",
			wantErr:  false,
		},
		{
			name:     "Invalid auth data",
			authData: "{invalid}",
			wantErr:  true,
		},
		{
			name:     "Valid auth data",
			authData: `{"username": "test", "password": "test123"}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ins := &Instance{
				AuthData: tt.authData,
			}
			err := ins.Decode()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.authData != "" {
					assert.NotEmpty(t, ins.AuthInfo)
				}
			}
		})
	}
}
