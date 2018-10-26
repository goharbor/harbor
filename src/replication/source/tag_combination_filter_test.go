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
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/models"
	"github.com/stretchr/testify/assert"

	"testing"
)

var tcfilter = NewTagCombinationFilter()

func TestTagCombinationFilterInit(t *testing.T) {
	assert.Nil(t, tcfilter.Init())
}

func TestTagCombinationFilterGetConverter(t *testing.T) {
	assert.Nil(t, tcfilter.GetConverter())
}

func TestTagCombinationFilterDoFilter(t *testing.T) {
	items := []models.FilterItem{
		{
			Kind: replication.FilterItemKindProject,
		},
		{
			Kind: replication.FilterItemKindRepository,
		},
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/ubuntu:invalid_tag:latest",
		},
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/ubuntu:14.04",
		},
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/ubuntu:16.04",
		},
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/centos:7",
		},
	}
	result := tcfilter.DoFilter(items)
	assert.Equal(t, 2, len(result))

	var ubuntu, centos models.FilterItem
	if result[0].Value == "library/ubuntu" {
		ubuntu = result[0]
		centos = result[1]
	} else {
		centos = result[0]
		ubuntu = result[1]
	}

	assert.Equal(t, replication.FilterItemKindRepository, ubuntu.Kind)
	assert.Equal(t, "library/ubuntu", ubuntu.Value)
	metadata, ok := ubuntu.Metadata["tags"].([]string)
	assert.True(t, ok)
	assert.EqualValues(t, []string{"14.04", "16.04"}, metadata)

	assert.Equal(t, replication.FilterItemKindRepository, centos.Kind)
	assert.Equal(t, "library/centos", centos.Value)
	metadata, ok = centos.Metadata["tags"].([]string)
	assert.True(t, ok)
	assert.EqualValues(t, []string{"7"}, metadata)
}
