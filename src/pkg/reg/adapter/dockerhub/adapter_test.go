package dockerhub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const (
	testUser     = ""
	testPassword = ""
)

// func mockRequest() *gock.Request {
// 	return gock.New("https://registry-1.docker.io")
// }

func getMockAdapter(t *testing.T) *adapter {
	r := &model.Registry{
		Type: model.RegistryTypeDockerHub,
		URL:  "https://registry-1.docker.io",
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

	// mockRequest().Get("/v2/repositories/namespaces").
	// 	Reply(http.StatusOK).BodyString("{}")

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

	// mockRequest().Get("/v2/repositories/goharbor/").
	// 	Reply(http.StatusOK).BodyString("{}")

	a := getMockAdapter(t)
	_, err := a.FetchArtifacts([]*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "goharbor/harbor-core",
		},
	})
	require.Nil(t, err)
}

// Integration test for real Docker Hub repo using Bearer token flow
func TestIntegration_DockerHub_GetManifest(t *testing.T) {
	// These credentials must be valid for Docker Hub
	const (
		dhUser = "<your_dockerhub_username>"
		dhPat  = "<your_dockerhub_personal_access_token>"
	)
	repoList := []struct {
		repo string
		tag  string
	}{
		{"goharbor/test1", "1.0"},
		{"goharbor/test1", "2.0"},
		{"goharbor/test1", "3.0"},
		{"goharbor/test2", "1.0"},
		{"goharbor/test3", "1.0"},
		{"goharbor/test4", "1.0"},
	}

	r := &model.Registry{
		Type: model.RegistryTypeDockerHub,
		URL:  registryURL, // Use registry-1.docker.io for manifest/blobs
		Credential: &model.Credential{
			AccessKey:    dhUser,
			AccessSecret: dhPat,
		},
	}
	ad, err := newAdapter(r)
	require.NoError(t, err)
	a := ad.(*adapter)

	for _, entry := range repoList {
		path := fmt.Sprintf("/v2/%s/manifests/%s", entry.repo, entry.tag)
		resp, err := a.client.Do(http.MethodGet, path, nil)
		if !assert.NoError(t, err, "repo: %s, tag: %s", entry.repo, entry.tag) {
			continue
		}
		assert.Equal(t, 200, resp.StatusCode, "repo: %s, tag: %s", entry.repo, entry.tag)
		_ = resp.Body.Close()
	}
}

// Integration test: download a real layer from Docker Hub
func TestIntegration_DockerHub_DownloadLayer(t *testing.T) {
	const (
		dhUser = "<your_dockerhub_username>"
		dhPat  = "<your_dockerhub_personal_access_token>"
		repo   = "goharbor/test1"
		tag    = "1.0"
	)

	r := &model.Registry{
		Type: model.RegistryTypeDockerHub,
		URL:  "https://registry-1.docker.io",
		Credential: &model.Credential{
			AccessKey:    dhUser,
			AccessSecret: dhPat,
		},
	}
	ad, err := newAdapter(r)
	require.NoError(t, err)
	a := ad.(*adapter)

	// Stap 1: Haal manifest op
	manifestPath := fmt.Sprintf("/v2/%s/manifests/%s", repo, tag)
	resp, err := a.client.Do(http.MethodGet, manifestPath, nil)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	defer resp.Body.Close()

	var manifest struct {
		Layers []struct {
			Digest string `json:"digest"`
		} `json:"layers"`
	}
	err = json.NewDecoder(resp.Body).Decode(&manifest)
	require.NoError(t, err)
	require.NotEmpty(t, manifest.Layers)
	layerDigest := manifest.Layers[0].Digest
	require.NotEmpty(t, layerDigest)

	// Stap 2: Download de eerste layer
	layerPath := fmt.Sprintf("/v2/%s/blobs/%s", repo, layerDigest)
	layerResp, err := a.client.Do(http.MethodGet, layerPath, nil)
	require.NoError(t, err)
	require.Equal(t, 200, layerResp.StatusCode)
	defer layerResp.Body.Close()

	// Lees een stukje van de layer om te valideren dat er data is
	buf := make([]byte, 512)
	n, err := layerResp.Body.Read(buf)
	require.NoError(t, err)
	require.Greater(t, n, 0)
}
