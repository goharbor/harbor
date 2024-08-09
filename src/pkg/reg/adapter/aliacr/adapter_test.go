package aliacr

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

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
				if buf, e := io.ReadAll(&io.LimitedReader{R: r.Body, N: 80}); e == nil {
					fmt.Println("\t", string(buf))
				}
				w.WriteHeader(http.StatusOK)
			},
		},
	)

	registry := &model.Registry{
		Type: model.RegistryTypeAliAcr,
		URL:  server.URL,
	}
	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    "MockAccessKey",
			AccessSecret: "MockAccessSecret",
		}
	}
	return &adapter{
		Adapter:  native.NewAdapter(registry),
		registry: registry,
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

func Test_getRegistryURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			"empty url",
			"",
			"",
			true,
		},
		{
			"just return url",
			"https://cr.cn-hangzhou.aliyun.com",
			"https://cr.cn-hangzhou.aliyun.com",
			false,
		},
		{
			"change match url",
			"https://cr.cn-hangzhou.aliyuncs.com",
			"https://registry.cn-hangzhou.aliyuncs.com",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRegistryURL(tt.url)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseRegistryService(t *testing.T) {
	tests := []struct {
		name     string
		service  string
		wantInfo *registryServiceInfo
		wantErr  bool
	}{
		{
			"not acr Service",
			"otherregistry.cn-hangzhou:china",
			nil,
			true,
		},
		{
			"empty Service",
			"",
			nil,
			true,
		},
		{
			"acr ee service",
			"registry.aliyuncs.com:cn-hangzhou:china:cri-xxxxxxxxx",
			&registryServiceInfo{
				IsACREE:    true,
				RegionID:   "cn-hangzhou",
				InstanceID: "cri-xxxxxxxxx",
			},
			false,
		},
		{
			"acr service",
			"registry.aliyuncs.com:cn-hangzhou:26842",
			&registryServiceInfo{
				IsACREE:    false,
				RegionID:   "cn-hangzhou",
				InstanceID: "",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parseRegistryService(tt.service)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
			assert.Equal(t, tt.wantInfo, info)
		})
	}
}

func Test_adapter_FetchArtifacts(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	var filters = []*model.Filter{}
	var resources, err = a.FetchArtifacts(filters)
	assert.NotNil(t, err)
	assert.Nil(t, resources)
}
func Test_aliyunAuthCredential_isCacheTokenValid(t *testing.T) {
	type fields struct {
		cacheToken          *registryTemporaryToken
		cacheTokenExpiredAt time.Time
	}

	var nilTime time.Time
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"nil cacheTokenExpiredAt", fields{nil, nilTime}, false},
		{"nil cacheToken", fields{nil, time.Time{}}, false},
		{"expired", fields{&registryTemporaryToken{}, time.Now().AddDate(0, 0, -1)}, false},
		{"ok", fields{&registryTemporaryToken{}, time.Now().AddDate(0, 0, 1)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &aliyunAuthCredential{
				cacheToken:          tt.fields.cacheToken,
				cacheTokenExpiredAt: tt.fields.cacheTokenExpiredAt,
			}
			if got := a.isCacheTokenValid(); got != tt.want {
				fmt.Println(got)
				assert.Equal(t, got, tt.want)
			}
		})
	}

}
