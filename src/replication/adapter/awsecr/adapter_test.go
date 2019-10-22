package awsecr

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
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

	factory, err = adp.GetFactory(model.RegistryTypeAwsEcr)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeAwsEcr,
		Credential: &model.Credential{
			AccessKey:    "xxx",
			AccessSecret: "ppp",
		},
		URL: "https://api.ecr.test-region.amazonaws.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, adapter)

	adapter, err = newAdapter(&model.Registry{
		Type: model.RegistryTypeAwsEcr,
		Credential: &model.Credential{
			AccessKey:    "xxx",
			AccessSecret: "ppp",
		},
		URL: "https://123456.dkr.ecr.test-region.amazonaws.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, adapter)

	adapter, err = newAdapter(&model.Registry{
		Type: model.RegistryTypeAwsEcr,
		Credential: &model.Credential{
			AccessKey:    "xxx",
			AccessSecret: "ppp",
		},
	})
	assert.Nil(t, adapter)
	assert.NotNil(t, err)

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
			AccessKey:    "xxx",
			AccessSecret: "ppp",
		}
	}
	dockerRegistryAdapter, err := native.NewAdapter(registry)
	if err != nil {
		panic(err)
	}
	return &adapter{
		registry:      registry,
		Adapter:       dockerRegistryAdapter,
		region:        "test-region",
		forceEndpoint: &server.URL,
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
	resources, err := a.FetchImages([]*model.Filter{
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

func TestAwsAuthCredential_Modify(t *testing.T) {
	et := time.Now().Add(time.Second).Unix()
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodPost,
			Pattern: "/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Method, r.URL)
				if buf, e := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: 80}); e == nil {
					fmt.Println("\t", string(buf))
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`
{
   "authorizationData" : [
      {
         "expiresAt" : %d,
         "proxyEndpoint" : "https://12345.dkr.ecr.ap-northeast-1.amazonaws.com",
         "authorizationToken" : "QVdTOmV5SndZWGxzYjJGa0lqb2llRlJNTkdSbWMyZE5RM0pXYWtoVWRrdzRNVVJIT1d0NFNXRlJjVEpET0c5cVZVUlFWRUkxUVVoS1ZDOVJjbXBIV0d4RlN6ZFZlR0UxTnk5TVdVSXdSU3RyVlRBNVRrSnVXbmhoUVdKaFlVZzFOV3d3YzJ4RVNIcHdZVWRZWTA4dmVGbHFjakphV1VKaE1YUlVkMU5JV2xWU1UxbzNSaTlKTHpaMFlXaFVPV1pXTldoelRXcFZiQ3R1SzBndlptdEtWMmMxYW5wclJrTXpkRXgzWkd4MFdWaE1PREZzV1dGWGQzVjJkbG94YkZKbFVrRnBhbVZYU1cxRksyRk9WM3AzVm1jM1N6aFNTMmhvYzBkRlFXNXRRbEJ4WkRGTVNuRkpjR2hUTldaNmJrazFNWEpPWmtwNU1WUnRMMVZNVFZZMVNYVkJjV1ZHYzA5MFUycEhkRTlFWVhsdGNrVlFXamhYUTBkR05YRklTWFE0UmtSTWNGQllPWFZoYW14NmNrbENkamROVkRsVk1UWlpkVlJpZDFaSWJYRjZPRGQ0VDNKdGVIaFRSR0Z2TDNCVU5qUXhja2w1YkhwUFJHUmpUMEpWWVdGUmRsWnpjak5TUzFaaFpEUmhaVkJ3ZUNzMGVYa3dhR2ROZW5sd1RVWkdRMkV3ZEVveU5HeEVNVVpUTWtkVWFXRlhMemMyVlRoTE9WWndNMEZ6WjFWaU4zbHZZbmhaY2tNNGFqVTRiMlJ4WlVWV01GbEtMekJTWWxSU1FYTjJiM1JEV1VzcllrSlJMM3BOUldjelN6UnpNa0lySzBGclIyUTJNM2MyZEd0VUx6VnBVWFZYY1UxdGJXcGxZMVF2TlRGc04wRm9UMFJzYlRKME5rbzRUamhvTXpkbWJVOWlMMkpXWTA1a09GWTVTM0ptWmtGQk1HTllSVTE1UWk5T1RIcHlaMEpwWWtOUmFFdHNiSEJTYlU5YVNHOHpWemxpV1haTmMzcHJVM0Z0YkU5clpUQmxjbU5GVTAxS1oycHRNV1Z5TDBKMlJHbGxVVEJDTWpSWVpHRjNLMDlGWkUxeVExTlRORll2ZDFFM1dWUXJRVFY0SzJScWNHWmhiREI1UkN0YVltOXdUelF3YlZBNFpXSlpLMUkxYld4VGRDOU9NblpxVDA1clMxbE9aemh3WVUxbVVFVjVja3BXT0ZSME4zUlVPR2MzV0ZaS1RVOVJOSEpqUlVaV09HbHBSVE5LVFRGc1RqRXljSGxOVFVGbk5sbGtNM3A0UW1OWFZrWkhRM1ZuVEhZNU1DOVhRVVJtTlV4TWFHbHFXRTh6YzNFd1dVaGtWRUkzUjBObVdtaGlla3huWjJsT2FVRXZXa2tyU1hSWlZUQnFUR2xxVGxoV1ZEWmFiRGcwZVdzM2IzWTVOVnBhZVUxUVFteHdjVkJLYkVsbU0yWkdWamc0UXpJdmVtTnNSeXQxYW1kb1VYVkxkM0U1ZEROdlVGZ3piaXREUm5oMVFqTnVTREZDVURGVlVVbEtVbGx5UlRaRUt6TkJPRWM1UVVabVVIRkNkMVZUWmpCSE5qbHlhM056YlhKdU1XMTVUa3RWZFZCemRETkplREpaY214ckswZFBOazB4ZG5GU2JsSXJUVTFUWmxSMmFtRlFOMXBEVW5CQmVFWTJZeXN5VlZKWVJrdElObEkyVDNCcGJFSktRV3N4UkhBMGRFNVBiVzAyYzJsalRFWjBjek0zTm1OUGNWWTRUMjkwVldZeGVrRjVZVGgxTlc5VWRGUkRUemcxZVVKVWFXNXNkMVZ1WmpZcmNtOHljVXRoVUZGWFdVZzNhamhWWkVaS05EUXdMMHRzVEdwNVlXSmlia1ZJTjNsRVpGRnhXRnBQTkVNeFptRlNZeTgyYUVwdEsxQXZXSEJETXpaSE0zTk9iWGgySzJKQlJHUXlUakZVYm1JMVJFZElZVVJTY2tsYU1uWkNiMHBRUW5GYVUwbGhRazV4YkZWUldWQjNUVEpEVjJzdlVVRTRlRVJDTjNsRlIwTnFSWFJuUFQwaUxDSmtZWFJoYTJWNUlqb2lRVkZGUWtGSVowRk5aa3RFYkVsdmNFTTJlbk13WWsxa1VuSlpVMGhoTDBNek9XdERjbU5RT0d0V2NISkZPV1lyYTFGQlFVRklOSGRtUVZsS1MyOWFTV2gyWTA1QlVXTkhiMGM0ZDJKUlNVSkJSRUp2UW1kcmNXaHJhVWM1ZHpCQ1FuZEZkMGhuV1VwWlNWcEpRVmRWUkVKQlJYVk5Ra1ZGUkVVNVpESnBPVVJNVUZrek5Ga3JTMmRCU1VKRlNVRTNNVnA2T1c1eVQzSjNVMnBhUW1Wc1YyOTBNRUpwY0VwbVoyTkhhbU5FU3k5WVEwY3JNSGxvTDFFNVpuZ3pVemc0WjFVMFQxQkVabVpVV1d4UFRUQTVPSGhvUjJWWmJscEZRV3hyZUN0ek1EMGlMQ0oyWlhKemFXOXVJam9pTWlJc0luUjVjR1VpT2lKRVFWUkJYMHRGV1NJc0ltVjRjR2x5WVhScGIyNGlPakUxTmpFME9EVXpNemg5"
      }
   ]
}
`, et)))
			},
		},
	)
	defer server.Close()
	a, _ := NewAuth("test-region", "xxx", "ppp", true).(*awsAuthCredential)
	a.forceEndpoint = &server.URL
	req := httptest.NewRequest(http.MethodGet, "https://1234.dkr.ecr.test-region.amazonaws.com/v2/", nil)
	err := a.Modify(req)
	assert.Nil(t, err)
	err = a.Modify(req)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	err = a.Modify(req)
	assert.Nil(t, err)
}

var urlForBenchmark = []string{
	"https://1234.dkr.ecr.test-region.amazonaws.com/v2/",
	"https://api.ecr.test-region.amazonaws.com",
	"https://test-region.amazonaws.com",
}

func compileRegexpEveryTime(url string) (string, error) {
	rs := regexp.MustCompile(regionPattern).FindStringSubmatch(url)
	if rs == nil {
		return "", errors.New("Bad aws url")
	}
	return rs[1], nil
}

func BenchmarkGetRegion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, url := range urlForBenchmark {
			parseRegion(url)
		}
	}
}

func BenchmarkCompileRegexpEveryTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, url := range urlForBenchmark {
			compileRegexpEveryTime(url)
		}
	}
}
