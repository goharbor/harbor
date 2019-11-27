package dockerhub

import (
	"fmt"
	"testing"

	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUser     = ""
	testPassword = ""
)

func getAdapter(t *testing.T) adp.Adapter {
	assert := assert.New(t)
	factory, err := adp.GetFactory(model.RegistryTypeDockerHub)
	assert.Nil(err)
	assert.NotNil(factory)

	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeDockerHub,
		URL:  baseURL,
		Credential: &model.Credential{
			AccessKey:    testUser,
			AccessSecret: testPassword,
		},
	})
	assert.Nil(err)
	assert.NotNil(adapter)

	return adapter
}

func TestInfo(t *testing.T) {
	adapter := &adapter{}
	info, err := adapter.Info()
	require.Nil(t, err)
	require.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func TestListCandidateNamespaces(t *testing.T) {
	adapter := &adapter{}
	namespaces, err := adapter.listCandidateNamespaces("library/*")
	require.Nil(t, err)
	require.Equal(t, 1, len(namespaces))
	assert.Equal(t, "library", namespaces[0])
}
func TestListNamespaces(t *testing.T) {
	if testUser == "" {
		return
	}

	assert := assert.New(t)
	ad := getAdapter(t)
	adapter := ad.(*adapter)

	namespaces, err := adapter.listNamespaces()
	assert.Nil(err)
	for _, ns := range namespaces {
		fmt.Println(ns)
	}
}

func TestFetchImages(t *testing.T) {
	ad := getAdapter(t)
	adapter := ad.(*adapter)
	_, err := adapter.FetchImages([]*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "goharbor/harbor-core",
		},
	})
	require.Nil(t, err)
}
