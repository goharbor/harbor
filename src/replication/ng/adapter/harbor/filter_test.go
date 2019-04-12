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

package harbor

import (
	"testing"

	"github.com/goharbor/harbor/src/replication/ng/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	// nil filters
	item := &FilterItem{}
	match, err := item.Match(nil)
	require.Nil(t, err)
	assert.True(t, match)
	// contains filter whose value isn't string
	item = &FilterItem{}
	filters := []*model.Filter{
		{
			Type:  "test",
			Value: 1,
		},
	}
	match, err = item.Match(filters)
	require.NotNil(t, err)
	// both filters match
	item = &FilterItem{
		Value: "b/c",
	}
	filters = []*model.Filter{
		{
			Value: "b/*",
		},
		{
			Value: "*/c",
		},
	}
	match, err = item.Match(filters)
	require.Nil(t, err)
	assert.True(t, match)
	// one filter matches and the other one doesn't
	item = &FilterItem{
		Value: "b/c",
	}
	filters = []*model.Filter{
		{
			Value: "b/*",
		},
		{
			Value: "d",
		},
	}
	match, err = item.Match(filters)
	require.Nil(t, err)
	assert.False(t, match)
	// both filters don't match
	item = &FilterItem{
		Value: "b/c",
	}
	filters = []*model.Filter{
		{
			Value: "f",
		},
		{
			Value: "d",
		},
	}
	match, err = item.Match(filters)
	require.Nil(t, err)
	assert.False(t, match)
}
