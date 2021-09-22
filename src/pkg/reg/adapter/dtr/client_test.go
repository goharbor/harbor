package dtr

import (
	"net/http"
	"testing"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjects(t *testing.T) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodPost,
			Pattern: "/api/v0/repositories",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				w.Write([]byte(`{
  "enableManifestLists": true,
  "immutableTags": true,
  "longDescription": "string",
  "name": "mynamespace/myrepo",
  "scanOnPush": true,
  "shortDescription": "string",
  "tagLimit": 0,
  "visibility": "public"
}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/api/v0/repositories/mynamespace/myrepo/tags",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Next-Page-Start", "")
				w.Write([]byte(`[
  {
    "author": "string",
    "createdAt": "2020-02-06T03:51:34.138Z",
    "digest": "string",
    "hashMismatch": true,
    "inNotary": true,
    "manifest": {
      "architecture": "string",
      "author": "string",
      "configDigest": "string",
      "configMediaType": "string",
      "createdAt": "2020-02-06T03:51:34.138Z",
      "digest": "string",
      "dockerfile": [
        {
          "isEmpty": true,
          "layerDigest": "string",
          "line": "string",
          "mediaType": "string",
          "size": 0,
          "urls": [
            "string"
          ]
        }
      ],
      "mediaType": "string",
      "os": "string",
      "osVersion": "string",
      "size": 0
    },
    "mirroring": {
      "digest": "string",
      "mirroringPolicyID": "string",
      "remoteRepository": "string",
      "remoteTag": "string"
    },
    "name": "mytag",
    "promotion": {
      "promotionPolicyID": "string",
      "sourceRepository": "string",
      "sourceTag": "string",
      "string": "string"
    },
    "updatedAt": "2020-02-06T03:51:34.138Z"
  }
]`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/api/v0/repositories/mynamespace/missingimage/tags",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/api/v0/repositories",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Next-Page-Start", "")
				w.Write([]byte(`{
  "repositories": [
    {
      "enableManifestLists": true,
      "id": "string",
      "immutableTags": true,
      "longDescription": "string",
      "name": "myrepo",
      "namespace": "mynamespace",
      "namespaceType": "user",
      "pulls": 0,
      "pushes": 0,
      "scanOnPush": true,
      "shortDescription": "string",
      "tagLimit": 0,
      "visibility": "public"
    }
  ]
}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/enzi/v0/accounts",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Next-Page-Start", "")
				w.Write([]byte(`{
  "accounts": [
    {
      "fullName": "string",
      "id": "string",
      "isActive": true,
      "isAdmin": true,
      "isImported": true,
      "isOrg": true,
      "membersCount": 0,
      "name": "mynamespace",
      "teamsCount": 0
    }
  ],
  "nextPageStart": "string",
  "orgsCount": 0,
  "resourceCount": 0,
  "usersCount": 0
}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodPost,
			Pattern: "/enzi/v0/accounts",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				w.Write([]byte(`{
  "fullName": "string",
  "isActive": true,
  "isAdmin": false,
  "isOrg": true,
  "name": "mynamespace",
  "password": "string",
  "searchLDAP": false
}`))
			},
		})
	client := &Client{
		url:      server.URL,
		username: "test",
		client: common_http.NewClient(
			&http.Client{
				Transport: common_http.GetHTTPTransport(common_http.WithInsecure(true)),
			}),
	}

	repositories, e := client.getRepositories()
	require.Nil(t, e)
	assert.Equal(t, 1, len(repositories))
	assert.Equal(t, "mynamespace/myrepo", repositories[0].Name)

	namespaces, e := client.getNamespaces()
	require.Nil(t, e)
	assert.Equal(t, 1, len(namespaces))
	assert.Equal(t, "mynamespace", namespaces[0].Name)

	tags, e := client.getTags("mynamespace/myrepo")
	require.Nil(t, e)
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "mytag", tags[0])

	// List tags for missign image
	_, e = client.getTags("mynamespace/missingimage")
	require.NotNil(t, e)

	e = client.createRepository("mynamespace/myrepo")
	require.Nil(t, e)

	e = client.createNamespace("mynamespace")
	require.Nil(t, e)

}
