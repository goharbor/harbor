package dtr

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
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
	a, err := newAdapter(registry)

	assert.Nil(t, err)
	return a.(*adapter), server
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
