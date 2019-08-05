package quayio

import (
	"testing"

	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
)

func getMockAdapter(t *testing.T) adp.Adapter {
	factory, _ := adp.GetFactory(model.RegistryTypeQuayio)
	adapter, err := factory(&model.Registry{
		Type: model.RegistryTypeQuayio,
		URL:  "https://quay.io",
	})
	assert.Nil(t, err)
	return adapter
}
func TestAdapter_NewAdapter(t *testing.T) {
	factory, err := adp.GetFactory("BadName")
	assert.Nil(t, factory)
	assert.NotNil(t, err)

	factory, err = adp.GetFactory(model.RegistryTypeQuayio)
	assert.Nil(t, err)
	assert.NotNil(t, factory)
}

func TestAdapter_HealthCheck(t *testing.T) {
	health, err := getMockAdapter(t).HealthCheck()
	assert.Nil(t, err)
	assert.Equal(t, string(health), model.Healthy)
}

func TestAdapter_Info(t *testing.T) {
	info, err := getMockAdapter(t).Info()
	assert.Nil(t, err)
	t.Log(info)
}

func TestAdapter_PullManifests(t *testing.T) {
	quayAdapter := getMockAdapter(t)
	registry, _, err := quayAdapter.(*adapter).PullManifest("quay/busybox", "latest", []string{})
	assert.Nil(t, err)
	assert.NotNil(t, registry)
	t.Log(registry)
}
