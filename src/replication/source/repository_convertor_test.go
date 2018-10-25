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

package source

import (
	"testing"

	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/models"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryConvert(t *testing.T) {
	items := []models.FilterItem{
		{
			Kind:  replication.FilterItemKindProject,
			Value: "library",
		},
		{
			Kind: replication.FilterItemKindRepository,
		},
	}
	expected := []models.FilterItem{
		{
			Kind:  replication.FilterItemKindRepository,
			Value: "library/ubuntu",
		},
		{
			Kind:  replication.FilterItemKindRepository,
			Value: "library/centos",
		},
		{
			Kind: replication.FilterItemKindRepository,
		},
	}

	converter := NewRepositoryConverter(&fakeRegistryAdaptor{})
	assert.EqualValues(t, expected, converter.Convert(items))
}

type fakeRegistryAdaptor struct{}

func (f *fakeRegistryAdaptor) Kind() string {
	return "fake"
}

func (f *fakeRegistryAdaptor) GetNamespaces() []models.Namespace {
	return nil
}

func (f *fakeRegistryAdaptor) GetNamespace(name string) models.Namespace {
	return models.Namespace{}
}

func (f *fakeRegistryAdaptor) GetRepositories(namespace string) []models.Repository {
	return []models.Repository{
		{
			Name: "library/ubuntu",
		},
		{
			Name: "library/centos",
		},
	}
}

func (f *fakeRegistryAdaptor) GetRepository(name string, namespace string) models.Repository {
	return models.Repository{}
}

func (f *fakeRegistryAdaptor) GetTags(repositoryName string, namespace string) []models.Tag {
	return []models.Tag{
		{
			Name: "14.04",
		},
		{
			Name: "16.04",
		},
	}
}

func (f *fakeRegistryAdaptor) GetTag(name string, repositoryName string, namespace string) models.Tag {
	return models.Tag{}
}
