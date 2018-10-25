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

var pfilter = NewPatternFilter(replication.FilterItemKindTag, "library/ubuntu:release-*", nil)

func TestPatternFilterInit(t *testing.T) {
	assert.Nil(t, pfilter.Init())
}

func TestPatternFilterGetConverter(t *testing.T) {
	assert.Nil(t, pfilter.GetConverter())
}

func TestPatternFilterDoFilter(t *testing.T) {
	items := []models.FilterItem{
		{
			Kind: replication.FilterItemKindProject,
		},
		{
			Kind: replication.FilterItemKindRepository,
		},
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/ubuntu:release-14.04",
		},
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/ubuntu:release-16.04",
		},
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/ubuntu:test",
		},
	}
	result := pfilter.DoFilter(items)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, replication.FilterItemKindTag, result[0].Kind)
	assert.Equal(t, "library/ubuntu:release-14.04", result[0].Value)
	assert.Equal(t, replication.FilterItemKindTag, result[1].Kind)
	assert.Equal(t, "library/ubuntu:release-16.04", result[1].Value)

}
