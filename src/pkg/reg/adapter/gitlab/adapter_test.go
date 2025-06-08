package gitlab

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func mustWriteHTTPResponse(t *testing.T, w io.Writer, fixturePath string) {
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("error opening fixture file: %v", err)
	}

	if _, err = io.Copy(w, f); err != nil {
		t.Fatalf("error writing response: %v", err)
	}
}
func getServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Www-Authenticate", "Bearer realm=\"http://"+r.Host+"/jwt/auth\",service=\"container_registry\"")
		w.WriteHeader(http.StatusUnauthorized)
	})

	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		w.Header().Set("X-Next-Page", "")

		switch search {
		case "library/dev-docker", "library", "library/", "dev-docker/", "dev-docker":
			mustWriteHTTPResponse(t, w, "testdata/projects/dev-docker.json")
		case "", "library/dockers":
			mustWriteHTTPResponse(t, w, "testdata/projects/all.json")
		default:
			w.Write([]byte(`[]`))
		}
	})

	mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		reRepo := regexp.MustCompile(`/api/v4/projects/(\d+)/registry/repositories/?$`)
		if match := reRepo.FindStringSubmatch(path); match != nil {
			projectID := match[1]
			w.Header().Set("X-Next-Page", "")
			mustWriteHTTPResponse(t, w, "testdata/repositories/"+projectID+".json")
			return
		}

		reTags := regexp.MustCompile(`/api/v4/projects/(\d+)/registry/repositories/(\d+)/tags$`)
		if match := reTags.FindStringSubmatch(path); match != nil {
			repoID := match[2]
			w.Header().Set("X-Next-Page", "")
			mustWriteHTTPResponse(t, w, "testdata/tags/"+repoID+".json")
			return
		}

		reDeleteTag := regexp.MustCompile(`/api/v4/projects/(\d+)/registry/repositories/(\d+)/tags/([^/]+)$`)
		if match := reDeleteTag.FindStringSubmatch(path); match != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.NotFound(w, r)
	})

	return httptest.NewServer(mux)
}

func getAdapter(t *testing.T) adp.Adapter {
	assertions := assert.New(t)
	factory, err := adp.GetFactory(model.RegistryTypeGitLab)
	assertions.Nil(err)
	assertions.NotNil(factory)
	server := getServer(t)
	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeGitLab,
		URL:  server.URL,
		Credential: &model.Credential{
			AccessKey:    "test",
			AccessSecret: "test",
		},
	})
	assertions.Nil(err)
	assertions.NotNil(adapter)

	return adapter
}
func TestFetchImages(t *testing.T) {
	assertions := assert.New(t)

	ad := getAdapter(t)
	adapter := ad.(*adapter)
	templates := map[string]int{
		"library/dev-docker": 1,
		"library/*-docker":   1,
		"library/dev-*":      2,
		"*/dev-docker":       1,
		"library/*":          2,
	}
	for k, v := range templates {
		resources, err := adapter.FetchArtifacts([]*model.Filter{
			{
				Type:  model.FilterTypeName,
				Value: k,
			},
		})
		require.Nil(t, err, k, v)
		assertions.Len(resources, v, k, v)
	}

	resources, err := adapter.FetchArtifacts([]*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "library/dockers",
		},
		{
			Type:  model.FilterTypeTag,
			Value: "{late*,v2}",
		},
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(resources))
	require.Equal(t, 2, len(resources[0].Metadata.Vtags))
}

func TestDeleteManifest(t *testing.T) {
	assertions := assert.New(t)
	ad := getAdapter(t)
	adapter := ad.(*adapter)

	t.Run("successful deletion", func(t *testing.T) {
		err := adapter.DeleteManifest("library/dockers", "harbor")
		require.Nil(t, err)
	})

	t.Run("successful deletion 1 level nested registry", func(t *testing.T) {
		err := adapter.DeleteManifest("library/dockers/harbor", "v0.1.1")
		require.Nil(t, err)
	})

	t.Run("successful deletion with dev-docker", func(t *testing.T) {
		err := adapter.DeleteManifest("library/dev-docker", "latest")
		require.Nil(t, err)
	})

	t.Run("no projects found", func(t *testing.T) {
		err := adapter.DeleteManifest("nonexistent/repository", "v1.0.0")
		require.NotNil(t, err)
		assertions.Contains(err.Error(), "no projects found")
	})

	t.Run("no repository found in project", func(t *testing.T) {
		err := adapter.DeleteManifest("library/nonexistent-repo", "v1.0.0")
		require.NotNil(t, err)
		assertions.Contains(err.Error(), "no projects found")
	})
}
