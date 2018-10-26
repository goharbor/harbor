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

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/replication"

	"github.com/goharbor/harbor/src/replication/models"
	"github.com/stretchr/testify/assert"
)

func TestInitOfLabelFilter(t *testing.T) {
	filter := NewLabelFilter(1)
	assert.Nil(t, filter.Init())
}

func TestGetConverterOfLabelFilter(t *testing.T) {
	filter := NewLabelFilter(1)
	assert.Nil(t, filter.GetConverter())
}

func TestDoFilterOfLabelFilter(t *testing.T) {
	test.InitDatabaseFromEnv()
	filter := NewLabelFilter(1)
	items := []models.FilterItem{
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/hello-world:latest",
		},
	}
	result := filter.DoFilter(items)
	assert.Equal(t, 0, len(result))
}
