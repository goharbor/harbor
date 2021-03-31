package googlegcr

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/test"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getMockAdapter(t *testing.T, hasCred, health bool) (*adapter, *httptest.Server) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/_catalog",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`
		{
			"repositories": [
					"test1"
			]
		}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/{repo}/tags/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`
			{
			    "name": "test1",
			    "tags": [
			        "latest"
			    ]
			}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Method, r.URL)
				if health {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Method, r.URL)
				w.WriteHeader(http.StatusOK)
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodPost,
			Pattern: "/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Method, r.URL)
				if buf, e := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: 80}); e == nil {
					fmt.Println("\t", string(buf))
				}
				w.WriteHeader(http.StatusOK)
			},
		},
	)
	registry := &model.Registry{
		Type: model.RegistryTypeGoogleGcr,
		URL:  server.URL,
	}
	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    "_json_key",
			AccessSecret: "ppp",
		}
	}

	factory, err := adp.GetFactory(model.RegistryTypeGoogleGcr)
	assert.Nil(t, err)
	assert.NotNil(t, factory)
	return newAdapter(registry), server
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

func TestAdapter_HealthCheck(t *testing.T) {
	a, s := getMockAdapter(t, false, true)
	defer s.Close()
	status, err := a.HealthCheck()
	assert.Nil(t, err)
	assert.NotNil(t, status)
	assert.EqualValues(t, model.Unhealthy, status)
	a, s = getMockAdapter(t, true, false)
	defer s.Close()
	status, err = a.HealthCheck()
	assert.Nil(t, err)
	assert.NotNil(t, status)
	assert.EqualValues(t, model.Unhealthy, status)
	a, s = getMockAdapter(t, true, true)
	defer s.Close()
	status, err = a.HealthCheck()
	assert.Nil(t, err)
	assert.NotNil(t, status)
	assert.EqualValues(t, model.Healthy, status)
}

func TestAdapter_PrepareForPush(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "busybox",
				},
			},
		},
	}
	err := a.PrepareForPush(resources)
	assert.Nil(t, err)
}

func TestAdapter_FetchImages(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	resources, err := a.FetchArtifacts([]*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "*",
		},
		{
			Type:  model.FilterTypeTag,
			Value: "*",
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, resources)
	assert.Equal(t, 1, len(resources))
}
