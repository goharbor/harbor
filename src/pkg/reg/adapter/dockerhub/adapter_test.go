package dockerhub

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
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

func TestConcurrentAdapterCreationDoesNotAuthenticate(testingContext *testing.T) {
	restoreNetwork := blockClientNetwork(testingContext)
	defer restoreNetwork()
	registry := &model.Registry{
		URL:        baseURL,
		Credential: &model.Credential{AccessKey: "test-user", AccessSecret: "test-secret"},
	}
	var workers sync.WaitGroup
	start := make(chan struct{})
	results := make(chan error, 32)
	for index := 0; index < cap(results); index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			_, err := newAdapter(registry)
			results <- err
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	for err := range results {
		require.NoError(testingContext, err)
	}
}

func TestNativePullDoesNotRequireDockerHubLogin(testingContext *testing.T) {
	manifestPayload := `{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[]}`
	manifestDigest := digest.FromString(manifestPayload).String()
	var hubLogins, registryLogins atomic.Int32
	var registryEndpoint string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/token" {
			username, secret, ok := request.BasicAuth()
			if !ok || username != "test-user" || secret != "valid-secret" {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}
			registryLogins.Add(1)
			_, _ = io.WriteString(writer, `{"token":"registry-bearer","expires_in":300}`)
			return
		}
		if request.Header.Get("Authorization") != "Bearer registry-bearer" {
			writer.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/token",service="test-registry"`, registryEndpoint))
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		if request.URL.Path == "/v2/library/busybox/manifests/latest" {
			writer.Header().Set("Docker-Content-Digest", manifestDigest)
			writer.Header().Set("Content-Type", "application/vnd.oci.image.index.v1+json")
			writer.Header().Set("Content-Length", fmt.Sprint(len(manifestPayload)))
			if request.Method == http.MethodGet {
				_, _ = io.WriteString(writer, manifestPayload)
			}
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	registryEndpoint = server.URL
	for _, secret := range []string{"valid-secret", "invalid-secret"} {
		testingContext.Run(secret, func(testingContext *testing.T) {
			registry := &model.Registry{URL: baseURL, Credential: &model.Credential{AccessKey: "test-user", AccessSecret: secret}}
			restoreNetwork := blockClientNetwork(testingContext)
			created, err := newAdapter(registry)
			restoreNetwork()
			require.NoError(testingContext, err)
			adapter := created.(*adapter)
			adapter.client.client.Transport = clientTestTransport(func(request *http.Request) (*http.Response, error) {
				hubLogins.Add(1)
				return clientTestResponse(http.StatusTooManyRequests, "{}"), nil
			})
			adapter.Adapter = native.NewAdapter(&model.Registry{URL: registryEndpoint, Credential: registry.Credential})
			health, err := adapter.HealthCheck()
			require.NoError(testingContext, err)
			if secret == "invalid-secret" {
				require.Equal(testingContext, model.Unhealthy, health)
				_, _, err = adapter.PullManifest("library/busybox", "latest")
				require.Error(testingContext, err)
				return
			}
			require.Equal(testingContext, model.Healthy, health)
			exists, descriptor, err := adapter.ManifestExist("library/busybox", "latest")
			require.NoError(testingContext, err)
			require.True(testingContext, exists)
			require.Equal(testingContext, manifestDigest, descriptor.Digest.String())
			manifest, pulledDigest, err := adapter.PullManifest("library/busybox", "latest")
			require.NoError(testingContext, err)
			require.NotNil(testingContext, manifest)
			require.Equal(testingContext, manifestDigest, pulledDigest)
		})
	}
	require.Zero(testingContext, hubLogins.Load())
	require.Positive(testingContext, registryLogins.Load())
}
