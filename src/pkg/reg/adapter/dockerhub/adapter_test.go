package dockerhub

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const (
	testUser     = ""
	testPassword = ""
)

func mockRequest() *gock.Request {
	return gock.New("https://hub.docker.com")
}

func getMockAdapter(t *testing.T) *adapter {
	r := &model.Registry{
		Type: model.RegistryTypeDockerHub,
		URL:  baseURL,
		Credential: &model.Credential{
			AccessKey:    testUser,
			AccessSecret: testPassword,
		},
	}
	ad, err := newAdapter(r)
	if err != nil {
		t.Fatalf("Failed to call newAdapter(), reason=[%v]", err)
	}
	a := ad.(*adapter)
	gock.InterceptClient(a.client.client)
	return a
}

func TestInfo(t *testing.T) {
	adapter := &adapter{}
	info, err := adapter.Info()
	require.Nil(t, err)
	require.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
	assert.Equal(t, model.RepositoryPathComponentTypeOnlyTwo, info.SupportedRepositoryPathComponentType)
}

func TestListCandidateNamespaces(t *testing.T) {
	adapter := &adapter{}
	namespaces, err := adapter.listCandidateNamespaces("library/*")
	require.Nil(t, err)
	require.Equal(t, 1, len(namespaces))
	assert.Equal(t, "library", namespaces[0])
}

func TestListNamespaces(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	mockRequest().Get("/v2/repositories/namespaces").
		Reply(http.StatusOK).BodyString("{}")

	a := getMockAdapter(t)

	namespaces, err := a.listNamespaces()
	assert.Nil(t, err)
	for _, ns := range namespaces {
		fmt.Println(ns)
	}
}

func TestFetchArtifacts(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	mockRequest().Get("/v2/repositories/goharbor/").
		Reply(http.StatusOK).BodyString("{}")

	a := getMockAdapter(t)
	_, err := a.FetchArtifacts([]*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "goharbor/harbor-core",
		},
	})
	require.Nil(t, err)
}
