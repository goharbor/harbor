// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flow

import (
	"testing"

	"github.com/stretchr/testify/suite"

	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/testing/mock"
)

type stageTestSuite struct {
	suite.Suite
}

func (s *stageTestSuite) SetupTest() {
}

func (s *stageTestSuite) TestInitialize() {
	factory := &mockFactory{}
	factory.On("AdapterPattern").Return(nil)
	adapter.RegisterFactory(model.RegistryTypeHarbor, factory)

	policy := &repctlmodel.Policy{
		SrcRegistry: &model.Registry{
			Type: model.RegistryTypeHarbor,
		},
		DestRegistry: &model.Registry{
			Type: model.RegistryTypeHarbor,
		},
	}
	factory.On("Create", mock.Anything).Return(&mockAdapter{}, nil)
	_, _, err := initialize(policy)
	s.Nil(err)
	factory.AssertExpectations(s.T())
}

func (s *stageTestSuite) TestFetchResources() {
	adapter := &mockAdapter{}
	adapter.On("FetchArtifacts", mock.Anything).Return([]*model.Resource{
		{},
		{},
	}, nil)
	policy := &repctlmodel.Policy{}
	resources, err := fetchResources(adapter, policy)
	s.Require().Nil(err)
	s.Len(resources, 2)
	adapter.AssertExpectations(s.T())
}

func (s *stageTestSuite) TestAssembleSourceResources() {
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeArtifact,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}
	policy := &repctlmodel.Policy{
		SrcRegistry: &model.Registry{
			ID: 1,
		},
	}
	res := assembleSourceResources(resources, policy)
	s.Len(res, 1)
	s.Equal(int64(1), res[0].Registry.ID)
}

func (s *stageTestSuite) TestAssembleDestinationResources() {
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeArtifact,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}
	policy := &repctlmodel.Policy{
		DestRegistry:              &model.Registry{},
		DestNamespace:             "test",
		DestNamespaceReplaceCount: -1,
		Override:                  true,
	}
	res, err := assembleDestinationResources(resources, policy, "")
	s.Require().Nil(err)
	s.Len(res, 1)
	s.Equal(model.ResourceTypeArtifact, res[0].Type)
	s.Equal("test/hello-world", res[0].Metadata.Repository.Name)
	s.Equal(1, len(res[0].Metadata.Vtags))
	s.Equal("latest", res[0].Metadata.Vtags[0])
}

func (s *stageTestSuite) TestReplaceNamespace() {
	// replace count <0, repository contains no "/"
	var (
		repository   string = "c"
		namespace    string = "n"
		replaceCount int8   = -1
	)
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("n/c", result)

	// replace count <0, repository contains only one "/"
	repository = "b/c"
	namespace = "n"
	replaceCount = -1
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("n/c", result)

	// replace count <0, repository contains more than one "/"
	repository = "a/b/c"
	namespace = "n"
	replaceCount = -1
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("n/c", result)

	// replace count > actual sub strings
	repository = "a/b"
	namespace = "n"
	replaceCount = 2
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().NotNil(err)

	// replace count = 0
	repository = "a/b/c"
	namespace = "n"
	replaceCount = 0
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("n/a/b/c", result)

	// replace count = 1
	repository = "a/b/c"
	namespace = "n"
	replaceCount = 1
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("n/b/c", result)

	// replace count = 2
	repository = "a/b/c"
	namespace = "n"
	replaceCount = 2
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("n/c", result)

	// the generated destination namespace contains 3 path component, but the destination registry requires only 2
	repository = "a/b/c"
	namespace = "n"
	replaceCount = 1
	result, err = replaceNamespace(repository, namespace, replaceCount, model.RepositoryPathComponentTypeOnlyTwo)
	s.Require().NotNil(err)

	// replace count =0, repository contains no "/"
	repository = "a"
	namespace = "n"
	replaceCount = 0
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("n/a", result)

	// replace count = 0 with empty namespace
	repository = "a/b/c"
	namespace = ""
	replaceCount = 0
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("a/b/c", result)
	
	// replace count = 1 with empty namespace
	repository = "a/b/c"
	namespace = ""
	replaceCount = 1
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("b/c", result)

	// replace count = 2 with empty namespace
	repository = "a/b/c"
	namespace = ""
	replaceCount = 2
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("c", result)

	// replace count <0, repository contains no "/" with empty namespace
	repository = "c"
	namespace = ""
	replaceCount = -1
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("c", result)

	// replace count <0, repository contains only one "/" with empty namespace
	repository = "b/c"
	namespace = ""
	replaceCount = -1
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("c", result)

	// replace count <0, repository contains more than one "/" with empty namespace
	repository = "a/b/c"
	namespace = ""
	replaceCount = -1
	result, err = replaceNamespace(repository, namespace, replaceCount, "")
	s.Require().Nil(err)
	s.Equal("c", result)
}

func TestStage(t *testing.T) {
	suite.Run(t, &stageTestSuite{})
}
