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
			got, err := newAdapter(tt.registry)
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
	var adapter = native{
		DefaultImageRegistry: reg,
		registry:             registry,
	}
	assert.NotNil(t, adapter)

	var info, err = adapter.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, registryTypeNative, info.Type)
	assert.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, 2, len(info.SupportedResourceFilters))
	assert.Equal(t, 2, len(info.SupportedTriggers))
	assert.Equal(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func Test_native_PrepareForPush(t *testing.T) {
	var registry = &model.Registry{URL: "abc"}
	var reg, _ = adp.NewDefaultImageRegistry(registry)
	var adapter = native{
		DefaultImageRegistry: reg,
		registry:             registry,
	}
	assert.NotNil(t, adapter)

	var err = adapter.PrepareForPush(nil)
	assert.Nil(t, err)
}
