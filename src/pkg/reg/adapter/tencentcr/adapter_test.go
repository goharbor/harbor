package tencentcr

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
)

var (
	mockAccessKey    = "AKIDxxxx"
	mockAccessSecret = "xxxxx"
	tcrClient        *tcr.Client
)

func setup() {
	os.Setenv("UTTEST", "true")

	if ak := os.Getenv("TENCENT_AK"); ak != "" {
		log.Info("USE AK from ENV")
		mockAccessKey = ak
	}
	if sk := os.Getenv("TENCENT_SK"); sk != "" {
		log.Info("USE SK from ENV")
		mockAccessSecret = sk
	}
	// var tcrCredential = common.NewCredential(mockAccessKey, mockAccessSecret)
	// var cfp = profile.NewClientProfile()

	// tcrClient, _ = tcr.NewClient(tcrCredential, regions.Guangzhou, cfp)
}

func teardown() {}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func TestAdapter_NewAdapter(t *testing.T) {
	factory, err := adp.GetFactory("BadName")
	assert.Nil(t, factory)
	assert.NotNil(t, err)

	factory, err = adp.GetFactory(model.RegistryTypeTencentTcr)
	assert.Nil(t, err)
	assert.NotNil(t, factory)
}

func TestAdapter_NewAdapter_NilAKSK(t *testing.T) {
	// Nil AK/SK
	adapter, err := newAdapter(&model.Registry{
		Type:       model.RegistryTypeTencentTcr,
		Credential: &model.Credential{},
	})
	assert.NotNil(t, err)
	assert.Nil(t, adapter)
}

func TestAdapter_NewAdapter_InvalidEndpoint(t *testing.T) {
	res := os.Getenv("UTTEST")
	os.Unsetenv("UTTEST")
	defer os.Setenv("UTTEST", res)

	// Invaild endpoint
	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeTencentTcr,
		Credential: &model.Credential{
			AccessKey:    mockAccessKey,
			AccessSecret: mockAccessSecret,
		},
		URL: "$$$",
	})
	assert.NotNil(t, err)
	assert.EqualError(t, err, errInvalidTcrEndpoint.Error())
	assert.Nil(t, adapter)
}

func TestAdapter_NewAdapter_Pingfailed(t *testing.T) {
	// Invaild endpoint
	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeTencentTcr,
		Credential: &model.Credential{
			AccessKey:    mockAccessKey,
			AccessSecret: mockAccessSecret,
		},
		URL: "https://.tencentcloudcr.com",
	})
	assert.NotNil(t, err)
	assert.Nil(t, adapter)
}

func TestAdapter_NewAdapter_InvalidAKSK(t *testing.T) {
	// Error AK/SK
	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeTencentTcr,
		Credential: &model.Credential{
			AccessKey:    "mockAccessKey",
			AccessSecret: "mockAccessSecret",
		},
	})
	assert.NotNil(t, err)
	assert.Nil(t, adapter)
}

func getTestServer() *httptest.Server {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Www-Authenticate", `Bearer realm="https://harbor-community.tencentcloudcr.com/service/token",service="harbor-registry"`)
				w.WriteHeader(http.StatusUnauthorized)
			},
		},
	)

	return server
}

func TestAdapter_NewAdapter_Ok(t *testing.T) {
	server := getTestServer()
	defer server.Close()

	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeTencentTcr,
		Credential: &model.Credential{
			AccessKey:    mockAccessKey,
			AccessSecret: mockAccessSecret,
		},
		URL: server.URL,
	})
	if sdkerr, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Infof("sdk error, error=%v", sdkerr)
		return
	}
	assert.NotNil(t, adapter)
	assert.Nil(t, err)

}

func TestAdapter_NewAdapter_InsecureOk(t *testing.T) {
	server := getTestServer()
	defer server.Close()

	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeTencentTcr,
		Credential: &model.Credential{
			AccessKey:    mockAccessKey,
			AccessSecret: mockAccessSecret,
		},
		Insecure: true,
		URL:      server.URL,
	})
	if sdkerr, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Infof("sdk error, error=%v", sdkerr)
		return
	}
	assert.NotNil(t, adapter)
	assert.Nil(t, err)
}

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
		Type: model.RegistryTypeAwsEcr,
		URL:  server.URL,
	}
	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    "AKIDxxxx",
			AccessSecret: "abcdefg",
		}
	}
	return &adapter{
		registry: registry,
		Adapter:  native.NewAdapter(registry),
	}, server
}

func TestAdapter_Info(t *testing.T) {
	tcrAdapter, _ := getMockAdapter(t, true, true)
	info, err := tcrAdapter.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)
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
	assert.NotNil(t, err)
}

func TestAdapter_PrepareForPush_NilResource(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	var resources = []*model.Resource{nil}

	err := a.PrepareForPush(resources)
	assert.NotNil(t, err)
}

func TestAdapter_PrepareForPush_NilMeata(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeImage,
		},
	}

	err := a.PrepareForPush(resources)
	assert.NotNil(t, err)
}

func TestAdapter_PrepareForPush_NilRepository(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	resources := []*model.Resource{
		{
			Type:     model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{},
		},
	}

	err := a.PrepareForPush(resources)
	assert.NotNil(t, err)
}

func TestAdapter_PrepareForPush_NilRepositoryName(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
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
	assert.NotNil(t, err)
}
