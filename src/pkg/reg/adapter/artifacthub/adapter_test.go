package artifacthub

import (
	"net/http"
	"testing"

	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func mockRequest() *gock.Request {
	return gock.New("https://artifacthub.io")
}

func getMockAdapter(t *testing.T) *adapter {
	ahRegistry := &model.Registry{
		Type: model.RegistryTypeArtifactHub,
		URL:  "https://artifacthub.io",
	}
	a, err := newAdapter(ahRegistry)
	if err != nil {
		t.Fatalf("Failed to call newAdapter(), reason=[%v]", err)
	}
	gock.InterceptClient(a.client.httpClient)
	return a
}

func TestAdapter_NewAdapter(t *testing.T) {
	factory, err := adp.GetFactory("BadName")
	assert.Nil(t, factory)
	assert.NotNil(t, err)

	factory, err = adp.GetFactory(model.RegistryTypeArtifactHub)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeArtifactHub,
		URL:  "https://artifacthub.io",
	})
	assert.Nil(t, err)
	assert.NotNil(t, adapter)
}

func TestAdapter_Info(t *testing.T) {
	a := getMockAdapter(t)
	info, err := a.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)

	assert.EqualValues(t, model.RegistryTypeArtifactHub, info.Type)
	assert.EqualValues(t, 1, len(info.SupportedResourceTypes))
	assert.EqualValues(t, model.ResourceTypeChart, info.SupportedResourceTypes[0])
}

func TestAdapter_HealthCheck(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	mockRequest().Get("/").Reply(http.StatusOK).BodyString("{}")

	a := getMockAdapter(t)
	h, err := a.HealthCheck()
	assert.Nil(t, err)
	assert.EqualValues(t, model.Healthy, h)
}

func TestAdapter_PrepareForPush(t *testing.T) {
	a := getMockAdapter(t)
	err := a.PrepareForPush(nil)
	assert.NotNil(t, err)
}

func TestAdapter_ChartExist(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	mockRequest().Get("/api/v1/packages/helm/harbor/harbor/1.5.0").
		Reply(http.StatusOK).BodyString("{}")
	mockRequest().Get("/api/v1/packages/helm/harbor/not-exists/1.5.0").
		Reply(http.StatusNotFound).BodyString("{}")
	mockRequest().Get("/api/v1/packages/helm/harbor/harbor/not-exists").
		Reply(http.StatusNotFound).BodyString("{}")

	a := getMockAdapter(t)

	b, err := a.ChartExist("harbor/harbor", "1.5.0")
	assert.Nil(t, err)
	assert.True(t, b)

	b, err = a.ChartExist("harbor/not-exists", "1.5.0")
	assert.Nil(t, err)
	assert.False(t, b)

	b, err = a.ChartExist("harbor/harbor", "not-exists")
	assert.Nil(t, err)
	assert.False(t, b)
}

func TestAdapter_DownloadChart(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	gock.New("https://helm.goharbor.io/").Get("/harbor-1.5.0.tgz").
		Reply(http.StatusOK).BodyString("{}")

	a := getMockAdapter(t)
	data, err := a.DownloadChart("harbor/harbor", "1.5.0", "")
	assert.NotNil(t, err)
	assert.Nil(t, data)

	data, err = a.DownloadChart("harbor/harbor", "1.5.0", "https://helm.goharbor.io/harbor-1.5.0.tgz")
	assert.Nil(t, err)
	assert.NotNil(t, data)
}

func TestAdapter_DeleteChart(t *testing.T) {
	a := getMockAdapter(t)

	err := a.DeleteChart("harbor/harbor", "1.5.0")
	assert.NotNil(t, err)
}

func TestAdapter_UploadChart(t *testing.T) {
	a := getMockAdapter(t)

	err := a.UploadChart("harbor/harbor", "1.5.0", nil)
	assert.NotNil(t, err)
}
