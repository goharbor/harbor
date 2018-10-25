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
	"github.com/goharbor/harbor/src/replication/registry"
	"github.com/stretchr/testify/assert"
)

func TestInitOfRepositoryFilter(t *testing.T) {
	filter := NewRepositoryFilter("", &registry.HarborAdaptor{})
	assert.Nil(t, filter.Init())
}

func TestGetConverterOfRepositoryFilter(t *testing.T) {
	filter := NewRepositoryFilter("", &registry.HarborAdaptor{})
	assert.NotNil(t, filter.GetConverter())
}

func TestDoFilterOfRepositoryFilter(t *testing.T) {
	// invalid filter item type
	filter := NewRepositoryFilter("", &registry.HarborAdaptor{})
	items := filter.DoFilter([]models.FilterItem{
		{
			Kind: "invalid_type",
		},
	})
	assert.Equal(t, 0, len(items))

	// empty pattern
	filter = NewRepositoryFilter("", &registry.HarborAdaptor{})
	items = filter.DoFilter([]models.FilterItem{
		{
			Kind:  replication.FilterItemKindRepository,
			Value: "library/hello-world",
		},
	})
	assert.Equal(t, 1, len(items))

	// non-empty pattern
	filter = NewRepositoryFilter("*", &registry.HarborAdaptor{})
	items = filter.DoFilter([]models.FilterItem{
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/hello-world",
		},
	})
	assert.Equal(t, 1, len(items))

	// non-empty pattern
	filter = NewRepositoryFilter("*", &registry.HarborAdaptor{})
	items = filter.DoFilter([]models.FilterItem{
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/hello-world:latest",
		},
	})
	assert.Equal(t, 1, len(items))
}
