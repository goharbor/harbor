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

package handler

import (
	"testing"

	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	api := &projectMetadataAPI{}

	tests := []struct {
		name      string
		metas     map[string]string
		expectErr bool
	}{
		{
			name:      "Invalid max upstream conn value",
			metas:     map[string]string{proModels.ProMetaMaxUpstreamConn: "invalid"},
			expectErr: true,
		},
		{
			name:      "max upstream conn value 0",
			metas:     map[string]string{proModels.ProMetaMaxUpstreamConn: "0"},
			expectErr: false,
		},
		{
			name:      "max upstream conn value -1",
			metas:     map[string]string{proModels.ProMetaMaxUpstreamConn: "-1"},
			expectErr: false,
		},
		{
			name:      "normal max upstream conn value",
			metas:     map[string]string{proModels.ProMetaMaxUpstreamConn: "30"},
			expectErr: false,
		},
		{
			name:      "Unsupported key",
			metas:     map[string]string{"unsupported_key": "value"},
			expectErr: true,
		},
		{
			name:      "Empty map",
			metas:     map[string]string{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := api.validate(tt.metas)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
