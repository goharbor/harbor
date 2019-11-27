package jfrog

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
)

const (
	fakeUploadID = "ac5fbe00-15f7-4d36-aa0e-cbdcdb15ec75"
	fakeDigest   = "sha256:f0f53b24e58a432aaa333d9993240340"

	fakeNamespace  = "mydocker"
	fakeRepository = "mydocker/nginx"
)

func getMockAdapter(t *testing.T, hasCred, health bool) (*adapter, *httptest.Server) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/artifactory/api/repositories",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
    {
        "key": "cyzhang",
        "description": "",
        "type": "LOCAL",
        "url": "http://49.4.2.82:8081/artifactory/cyzhang",
        "packageType": "Docker"
    },
    {
        "key": "mydocker",
        "type": "LOCAL",
        "url": "http://49.4.2.82:8081/artifactory/mydocker",
        "packageType": "Docker"
    }
]`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodPut,
			Pattern: fmt.Sprintf("/artifactory/api/repositories/%s", fakeNamespace),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodPost,
			Pattern: fmt.Sprintf("/v2/%s/blobs/uploads/", fakeRepository),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Docker-Upload-Uuid", fakeUploadID)
				w.WriteHeader(http.StatusAccepted)
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodPatch,
			Pattern: fmt.Sprintf("/v2/%s/blobs/uploads/%s", fakeRepository, fakeUploadID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodPut,
			Pattern: fmt.Sprintf("/v2/%s/blobs/uploads/%s", fakeRepository, fakeUploadID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
		},
	)

	registry := &model.Registry{
		Type: model.RegistryTypeJfrogArtifactory,
		URL:  server.URL,
	}

	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    "admin",
			AccessSecret: "password",
		}
	}

	factory, err := adp.GetFactory(model.RegistryTypeJfrogArtifactory)
	assert.Nil(t, err)
	assert.NotNil(t, factory)
	a, err := newAdapter(registry)

	assert.Nil(t, err)
	return a.(*adapter), server
}

func TestAdapter_Info(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	info, err := a.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.EqualValues(t, 1, len(info.SupportedResourceTypes))
	assert.EqualValues(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func TestAdapter_PrepareForPush(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "mydocker/busybox",
				},
			},
		},
	}
	err := a.PrepareForPush(resources)
	assert.Nil(t, err)
}

func TestAdapter_PushBlob(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	err := a.PushBlob(fakeRepository, fakeDigest, 20, bytes.NewReader([]byte("test")))
	assert.Nil(t, err)
}
