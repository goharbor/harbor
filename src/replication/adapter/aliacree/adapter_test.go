package aliacree

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/utils/test"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"

	"github.com/stretchr/testify/assert"
)

func TestAdapter_NewAdapter(t *testing.T) {
	factory, err := adp.GetFactory("BadName")
	assert.Nil(t, factory)
	assert.NotNil(t, err)

	factory, err = adp.GetFactory(model.RegistryTypeAliAcrEE)
	assert.Nil(t, err)
	assert.NotNil(t, factory)
}

func getMockAdapter(t *testing.T, hasCred, health bool) (*adapter, *httptest.Server) {
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
		Type: model.RegistryTypeAliAcrEE,
		URL:  server.URL,
	}
	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    "MockAccessKey",
			AccessSecret: "MockAccessSecret",
		}
	}
	nativeRegistry := native.NewAdapter(registry)

	return &adapter{
		Adapter:   nativeRegistry,
		region:    "test-region",
		domain:    server.URL,
		accessKey: "MockAccessKey",
		secretKey: "MockAccessSecret",
		registry:  registry,
	}, server
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

func Test_getRegion(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantRegion   string
		wantInstance string
		wantErr      bool
	}{
		{"public registry shanghai enterprise", "https://foo-bar-registry.cn-shanghai.cr.aliyuncs.com", "cn-shanghai", "foo-bar", false},
		{"private registry shanghai enterprise", "https://foo-bar-registry-vpc.cn-shanghai.cr.aliyuncs.com", "cn-shanghai", "foo-bar", false},
		{"public registry shanghai enterprise with complex name", "https://foo1.bar-oof_0x-registry.cn-shanghai.cr.aliyuncs.com", "cn-shanghai", "foo1.bar-oof_0x", false},
		{"private registry shanghai enterprise with complex name", "https://foo1.bar-oof_0x-registry-vpc.cn-shanghai.cr.aliyuncs.com", "cn-shanghai", "foo1.bar-oof_0x", false},
		{"invalid public registry shanghai enterprise", "https://google.com", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInstance, gotRegion, err := getInstanceRegion(tt.url)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
			assert.Equal(t, tt.wantRegion, gotRegion)
			assert.Equal(t, tt.wantInstance, gotInstance)
		})
	}
}

var urlForBenchmark = []string{
	"https://cr.cn-hangzhou.aliyuncs.com",
	"https://registry.cn-shanghai.aliyuncs.com",
}

func BenchmarkGetRegion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, url := range urlForBenchmark {
			getInstanceRegion(url)
		}
	}
}

func Test_adapter_FetchImages(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	var filters = []*model.Filter{}
	var resources, err = a.FetchArtifacts(filters)
	assert.NotNil(t, err)
	assert.Nil(t, resources)
}

func Test_aliyunAuthCredential_isCacheTokenValid(t *testing.T) {
	type fields struct {
		region              string
		instanceID          string
		accessKey           string
		secretKey           string
		cacheToken          *registryTemporaryToken
		cacheTokenExpiredAt time.Time
	}

	var nilTime time.Time
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"nil cacheTokenExpiredAt", fields{"test-region", "test-instance-id", "MockAccessKey", "MockSecretKey", nil, nilTime}, false},
		{"nil cacheToken", fields{"test-region", "test-instance-id", "MockAccessKey", "MockSecretKey", nil, time.Time{}}, false},
		{"expired", fields{"test-region", "test-instance-id", "MockAccessKey", "MockSecretKey", &registryTemporaryToken{}, time.Now().AddDate(0, 0, -1)}, false},
		{"ok", fields{"test-region", "test-instance-id", "MockAccessKey", "MockSecretKey", &registryTemporaryToken{}, time.Now().AddDate(0, 0, 1)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &acrEEAuthCredential{
				region:              tt.fields.region,
				instanceID:          tt.fields.instanceID,
				accessKey:           tt.fields.accessKey,
				secretKey:           tt.fields.secretKey,
				cacheToken:          tt.fields.cacheToken,
				cacheTokenExpiredAt: tt.fields.cacheTokenExpiredAt,
			}
			if got := a.isCacheTokenValid(); got != tt.want {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}
