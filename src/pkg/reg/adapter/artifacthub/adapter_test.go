package artifacthub

import (
	"net/http"
	"os"
	"testing"

	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

var ahAdapter *adapter

func init() {
	var err error
	ahRegistry := &model.Registry{
		Type: model.RegistryTypeArtifactHub,
		URL:  "https://artifacthub.io",
	}
	ahAdapter, err = newAdapter(ahRegistry)
	if err != nil {
		os.Exit(1)
	}
	gock.InterceptClient(ahAdapter.client.httpClient)
}

func mockRequest() *gock.Request {
	return gock.New("https://artifacthub.io")
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
	info, err := ahAdapter.Info()
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

	h, err := ahAdapter.HealthCheck()
	assert.Nil(t, err)
	assert.EqualValues(t, model.Healthy, h)
}

func TestAdapter_PrepareForPush(t *testing.T) {
	err := ahAdapter.PrepareForPush(nil)
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

	b, err := ahAdapter.ChartExist("harbor/harbor", "1.5.0")
	assert.Nil(t, err)
	assert.True(t, b)

	b, err = ahAdapter.ChartExist("harbor/not-exists", "1.5.0")
	assert.Nil(t, err)
	assert.False(t, b)

	b, err = ahAdapter.ChartExist("harbor/harbor", "not-exists")
	assert.Nil(t, err)
	assert.False(t, b)
}

func TestAdapter_DownloadChart(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	gock.New("https://helm.goharbor.io/").Get("/harbor-1.5.0.tgz").
		Reply(http.StatusOK).BodyString("{}")

	data, err := ahAdapter.DownloadChart("harbor/harbor", "1.5.0", "")
	assert.NotNil(t, err)
	assert.Nil(t, data)

	data, err = ahAdapter.DownloadChart("harbor/harbor", "1.5.0", "https://helm.goharbor.io/harbor-1.5.0.tgz")
	assert.Nil(t, err)
	assert.NotNil(t, data)
}

func TestAdapter_DeleteChart(t *testing.T) {
	err := ahAdapter.DeleteChart("harbor/harbor", "1.5.0")
	assert.NotNil(t, err)
}

func TestAdapter_UploadChart(t *testing.T) {
	err := ahAdapter.UploadChart("harbor/harbor", "1.5.0", nil)
	assert.NotNil(t, err)
}
