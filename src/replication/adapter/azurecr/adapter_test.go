package azurecr

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/utils/test"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
)

func getMockAdapter(t *testing.T, hasCred, health bool) (*adapter, *httptest.Server) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/_catalog",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"repositories": ["test1"]}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/{repo}/tags/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"name": "test1", "tags": ["latest"]}`))
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
		Type: model.RegistryTypeAzureAcr,
		URL:  server.URL,
	}
	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    "acr",
			AccessSecret: "pwd",
		}
	}

	factory, err := adp.GetFactory(model.RegistryTypeAzureAcr)
	assert.Nil(t, err)
	assert.NotNil(t, factory)
	a, err := factory(registry)

	assert.Nil(t, err)
	return a.(*adapter), server
}

func TestInfo(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	info, err := a.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.EqualValues(t, 1, len(info.SupportedResourceTypes))
	assert.EqualValues(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func TestHealthCheck(t *testing.T) {
	a, s := getMockAdapter(t, true, false)
	defer s.Close()
	status, err := a.HealthCheck()
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

func TestPrepareForPush(t *testing.T) {
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
