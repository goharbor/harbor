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

package internal

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMethodsOfLink(t *testing.T) {
	str := `<http://example.com/TheBook/chapter2>; rel="previous"; title="previous chapter" , <http://example.com/TheBook/chapter4>; rel="next"; title="next chapter"`
	links := ParseLinks(str)
	require.Len(t, links, 2)
	assert.Equal(t, "http://example.com/TheBook/chapter2", links[0].URL)
	assert.Equal(t, "previous", links[0].Rel)
	assert.Equal(t, "previous chapter", links[0].Attrs["title"])
	assert.Equal(t, "http://example.com/TheBook/chapter4", links[1].URL)
	assert.Equal(t, "next", links[1].Rel)
	assert.Equal(t, "previous", links[0].Rel)
	assert.Equal(t, "next chapter", links[1].Attrs["title"])

	s := links.String()
	assert.Equal(t, str, s)
}
