package dockerhub

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	adp "github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/model"
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

	adapter, err := factory(&model.Registry{
		Type: model.RegistryTypeDockerHub,
		Credential: &model.Credential{
			AccessKey:    testUser,
			AccessSecret: testPassword,
		},
	})
	assert.Nil(err)
	assert.NotNil(adapter)

	return adapter
}

func TestListNamespaces(t *testing.T) {
	if testUser == "" {
		return
	}

	assert := assert.New(t)
	adapter := getAdapter(t)

	namespaces, err := adapter.ListNamespaces(nil)
	assert.Nil(err)
	for _, ns := range namespaces {
		fmt.Println(ns.Name)
	}
}

func TestCreateNamespace(t *testing.T) {
	if testUser == "" {
		return
	}

	assert := assert.New(t)
	adapter := getAdapter(t)

	err := adapter.CreateNamespace(&model.Namespace{
		Name: "harborns",
		Metadata: map[string]interface{}{
			metadataKeyFullName: "harbor namespace",
			metadataKeyCompany:  "harbor",
		},
	})
	assert.Nil(err)
}
