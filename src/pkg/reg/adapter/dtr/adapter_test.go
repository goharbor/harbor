package dtr

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	a := &adapter{}
	info, err := a.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.EqualValues(t, 1, len(info.SupportedResourceTypes))
	assert.EqualValues(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func getMockAdapter(t *testing.T, hasCred, health bool) (*adapter, *httptest.Server) {
	server := test.NewServer(
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
  },
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
    "name": "v1.0.0",
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
		})

	registry := &model.Registry{
		Type: model.RegistryTypeDTR,
		URL:  server.URL,
	}

	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    "admin",
			AccessSecret: "password",
		}
	}

	factory, err := adp.GetFactory(model.RegistryTypeDTR)
	assert.Nil(t, err)
	assert.NotNil(t, factory)
	a := newAdapter(registry)

	assert.Nil(t, err)
	return a, server
}

func TestAdapter_PrepareForPush(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "mynamespace/myrepo",
				},
			},
		},
	}
	err := a.PrepareForPush(resources)
	assert.Nil(t, err)
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

func TestAdapter_FetchArtifacts(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	filters := []*model.Filter{}
	r, err := a.FetchArtifacts(filters)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(r))
	assert.EqualValues(t, 2, len(r[0].Metadata.Artifacts))

}

func TestAdapter_FetchArtifactsFiltered(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()

	testCases := []struct {
		nameFilter string
		tagFilter  string
		repos      int
		artifacts  int
	}{
		{"mynamespace/**", "**", 1, 2},
		{"mynamespace/myrepo", "**", 1, 2},
		{"mynamespace/myrepo", "v1.0.0", 1, 1},
		{"mynamespace/myrepo", "notfound", 1, 0},
	}
	for _, tc := range testCases {

		filters := []*model.Filter{
			{
				Type:  model.FilterTypeName,
				Value: tc.nameFilter,
			},
			{
				Type:  model.FilterTypeTag,
				Value: tc.tagFilter,
			},
		}
		r, err := a.FetchArtifacts(filters)
		if err != nil {
			t.Fatalf("could fetch artifacts for repo=%q tag=%s", tc.nameFilter, tc.tagFilter)
		}
		if len(r) != tc.repos {
			t.Fatalf("wrong number of repos returned for repo=%q tag=%s, wanted %d got %d", tc.nameFilter, tc.tagFilter, tc.repos, len(r))
		}
		if len(r[0].Metadata.Artifacts) != tc.artifacts {
			t.Fatalf("wrong number of artifacts returned for repo=%q tag=%s, wanted %d got %d", tc.nameFilter, tc.tagFilter, tc.artifacts, len(r))
		}

	}

}
