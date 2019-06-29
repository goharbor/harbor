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

package native

import (
	"testing"

	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
)

func Test_newAdapter(t *testing.T) {
	tests := []struct {
		name     string
		registry *model.Registry
		wantErr  bool
	}{
		{name: "Nil Registry URL", registry: &model.Registry{}, wantErr: true},
		{name: "Right", registry: &model.Registry{URL: "abc"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAdapter(tt.registry)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, got)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func Test_native_Info(t *testing.T) {
	var registry = &model.Registry{URL: "abc"}
	var reg, _ = adp.NewDefaultImageRegistry(registry)
	var adapter = Native{
		DefaultImageRegistry: reg,
		registry:             registry,
	}
	assert.NotNil(t, adapter)

	var info, err = adapter.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, model.RegistryTypeDockerRegistry, info.Type)
	assert.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, 2, len(info.SupportedResourceFilters))
	assert.Equal(t, 2, len(info.SupportedTriggers))
	assert.Equal(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func Test_native_PrepareForPush(t *testing.T) {
	var registry = &model.Registry{URL: "abc"}
	var reg, _ = adp.NewDefaultImageRegistry(registry)
	var adapter = Native{
		DefaultImageRegistry: reg,
		registry:             registry,
	}
	assert.NotNil(t, adapter)

	var err = adapter.PrepareForPush(nil)
	assert.Nil(t, err)
}
