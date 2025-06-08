package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils/test"
)

func TestProjects(t *testing.T) {
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/v4/projects",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			data := `[

{
        "id": 12312344,
        "description": "",
        "name": "dockers",
        "name_with_namespace": "Library / dockers",
        "path": "dockers",
        "path_with_namespace": "library/dockers",
        "created_at": "2019-01-17T09:47:07.504Z",
        "default_branch": "master",
        "tag_list": [],
    
        "avatar_url": null,
        "star_count": 0,
        "forks_count": 0,
        "last_activity_at": "2019-06-09T15:18:10.045Z",
        "empty_repo": false,
        "archived": false,
        "visibility": "private",
        "resolve_outdated_diff_discussions": false,
        "container_registry_enabled": true,
        "issues_enabled": true,
        "merge_requests_enabled": true,
        "wiki_enabled": true,
        "jobs_enabled": true,
        "snippets_enabled": true,
        "shared_runners_enabled": true,
        "lfs_enabled": true,
        "creator_id": 123412412,
        "forked_from_project": {},
        "import_status": "finished",
        "open_issues_count": 0,
        "ci_default_git_depth": null,
        "public_jobs": true,
        "ci_config_path": null,
        "shared_with_groups": [],
        "only_allow_merge_if_pipeline_succeeds": false,
        "request_access_enabled": false,
        "only_allow_merge_if_all_discussions_are_resolved": false,
        "printing_merge_request_link_enabled": true,
        "merge_method": "merge",
        "external_authorization_classification_label": "",
        "permissions": {
            "project_access": null,
            "group_access": null
        },
        "mirror": false
    }

]`
			w.Header().Set("X-Next-Page", "")
			w.Write([]byte(data))
		},
	})
	client := &Client{
		url:      server.URL,
		username: "test",
		token:    "test",
		client: common_http.NewClient(
			&http.Client{
				Transport: common_http.GetHTTPTransport(common_http.WithInsecure(true)),
			}),
	}
	projects, e := client.getProjects()
	require.Nil(t, e)
	assert.Equal(t, 1, len(projects))
}

func TestDeleteTag(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := test.NewServer(&test.RequestHandlerMapping{
			Method:  http.MethodDelete,
			Pattern: "/api/v4/projects/123/registry/repositories/456/tags/v1.0.0",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		})
		defer server.Close()

		client := &Client{
			url:      server.URL,
			username: "test",
			token:    "test",
			client: common_http.NewClient(
				&http.Client{
					Transport: common_http.GetHTTPTransport(common_http.WithInsecure(true)),
				}),
		}

		err := client.deleteTag(123, 456, "v1.0.0")
		require.Nil(t, err)
	})

	t.Run("deletion with 204 status", func(t *testing.T) {
		server := test.NewServer(&test.RequestHandlerMapping{
			Method:  http.MethodDelete,
			Pattern: "/api/v4/projects/123/registry/repositories/456/tags/v1.0.0",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
		})
		defer server.Close()

		client := &Client{
			url:      server.URL,
			username: "test",
			token:    "test",
			client: common_http.NewClient(
				&http.Client{
					Transport: common_http.GetHTTPTransport(common_http.WithInsecure(true)),
				}),
		}

		err := client.deleteTag(123, 456, "v1.0.0")
		require.Nil(t, err)
	})

	t.Run("deletion with 404 not found", func(t *testing.T) {
		server := test.NewServer(&test.RequestHandlerMapping{
			Method:  http.MethodDelete,
			Pattern: "/api/v4/projects/123/registry/repositories/456/tags/v1.0.0",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "Tag not found"}`))
			},
		})
		defer server.Close()

		client := &Client{
			url:      server.URL,
			username: "test",
			token:    "test",
			client: common_http.NewClient(
				&http.Client{
					Transport: common_http.GetHTTPTransport(common_http.WithInsecure(true)),
				}),
		}

		err := client.deleteTag(123, 456, "v1.0.0")
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to delete tag with status code 404")
		assert.Contains(t, err.Error(), "Tag not found")
	})

	t.Run("deletion with 403 forbidden", func(t *testing.T) {
		server := test.NewServer(&test.RequestHandlerMapping{
			Method:  http.MethodDelete,
			Pattern: "/api/v4/projects/123/registry/repositories/456/tags/v1.0.0",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "Access denied"}`))
			},
		})
		defer server.Close()

		client := &Client{
			url:      server.URL,
			username: "test",
			token:    "test",
			client: common_http.NewClient(
				&http.Client{
					Transport: common_http.GetHTTPTransport(common_http.WithInsecure(true)),
				}),
		}

		err := client.deleteTag(123, 456, "v1.0.0")
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to delete tag with status code 403")
		assert.Contains(t, err.Error(), "Access denied")
	})

	t.Run("deletion with 500 internal server error", func(t *testing.T) {
		server := test.NewServer(&test.RequestHandlerMapping{
			Method:  http.MethodDelete,
			Pattern: "/api/v4/projects/123/registry/repositories/456/tags/v1.0.0",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message": "Internal server error"}`))
			},
		})
		defer server.Close()

		client := &Client{
			url:      server.URL,
			username: "test",
			token:    "test",
			client: common_http.NewClient(
				&http.Client{
					Transport: common_http.GetHTTPTransport(common_http.WithInsecure(true)),
				}),
		}

		err := client.deleteTag(123, 456, "v1.0.0")
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to delete tag with status code 500")
		assert.Contains(t, err.Error(), "Internal server error")
	})

	t.Run("tag name with special characters", func(t *testing.T) {
		server := test.NewServer(&test.RequestHandlerMapping{
			Method:  http.MethodDelete,
			Pattern: "/api/v4/projects/123/registry/repositories/456/tags/v1.0.0-alpha.1",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		})
		defer server.Close()

		client := &Client{
			url:      server.URL,
			username: "test",
			token:    "test",
			client: common_http.NewClient(
				&http.Client{
					Transport: common_http.GetHTTPTransport(common_http.WithInsecure(true)),
				}),
		}

		err := client.deleteTag(123, 456, "v1.0.0-alpha.1")
		require.Nil(t, err)
	})
}
