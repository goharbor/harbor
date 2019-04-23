package dockerhub

import (
	"fmt"
	"testing"

	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
)

// TODO add more unit test

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
	ad := getAdapter(t)
	adapter := ad.(*adapter)

	namespaces, err := adapter.listNamespaces()
	assert.Nil(err)
	for _, ns := range namespaces {
		fmt.Println(ns)
	}
}
