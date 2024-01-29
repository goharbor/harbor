package volcenginecr

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
	volcCR "github.com/volcengine/volcengine-go-sdk/service/cr"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	volcSession "github.com/volcengine/volcengine-go-sdk/volcengine/session"
)

func getMockAdapter_withoutCred(t *testing.T, hasCred, health bool) (*adapter, *httptest.Server) {
	server := test.NewServer(
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
		Type: model.RegistryTypeVolcCR,
		URL:  server.URL,
	}
	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    "MockAccessKey",
			AccessSecret: "MockAccessSecret",
		}
	}
	name := "test-registry"
	config := volcengine.NewConfig().
		WithCredentials(credentials.NewStaticCredentials("", "", "")).
		WithRegion("cn-beijing")
	sess, _ := volcSession.NewSession(config)
	client := volcCR.New(sess)
	return &adapter{
		Adapter:      native.NewAdapter(registry),
		registryName: &name,
		volcCrClient: client,
		registry:     registry,
	}, server
}

func TestAdapter_NewAdapter_InvalidURL(t *testing.T) {
	factory, err := adp.GetFactory("BadName")
	assert.Nil(t, factory)
	assert.Error(t, err)

	factory, err = adp.GetFactory(model.RegistryTypeVolcCR)
	assert.NoError(t, err)
	assert.NotNil(t, factory)
	adapter, err := factory.Create(&model.Registry{
		Type:       model.RegistryTypeVolcCR,
		Credential: &model.Credential{},
	})
	assert.Error(t, err)
	assert.Nil(t, adapter)
}

func TestAdapter_NewAdapter_PingFailed(t *testing.T) {
	factory, _ := adp.GetFactory(model.RegistryTypeVolcCR)
	adapter, err := factory.Create(&model.Registry{
		Type:       model.RegistryTypeVolcCR,
		Credential: &model.Credential{},
		URL:        "https://cr-test-cn-beijing.cr.volces.com",
	})
	assert.Error(t, err)
	assert.Nil(t, adapter)
}

func TestAdapter_Info(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	info, err := a.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)

	assert.EqualValues(t, 1, len(info.SupportedResourceTypes))
	assert.EqualValues(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func TestAdapter_PrepareForPush(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
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
	assert.Error(t, err)
}

func TestAdapter_PrepareForPush_NilResource(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	var resources = []*model.Resource{nil}

	err := a.PrepareForPush(resources)
	assert.Error(t, err)
}

func TestAdapter_PrepareForPush_NilMeta(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeImage,
		},
	}

	err := a.PrepareForPush(resources)
	assert.Error(t, err)
}

func TestAdapter_PrepareForPush_NilRepository(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	resources := []*model.Resource{
		{
			Type:     model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{},
		},
	}

	err := a.PrepareForPush(resources)
	assert.Error(t, err)
}

func TestAdapter_PrepareForPush_NilRepositoryName(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{},
			},
		},
	}

	err := a.PrepareForPush(resources)
	assert.Error(t, err)
}
