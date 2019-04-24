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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHTTPTransport(t *testing.T) {
	transport := GetHTTPTransport(true)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	transport = GetHTTPTransport(false)
	assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestParseRepository(t *testing.T) {
	// empty repository
	repository := ""
	namespace, rest := ParseRepository(repository)
	assert.Equal(t, "", namespace)
	assert.Equal(t, "", rest)
	// repository contains no "/"
	repository = "c"
	namespace, rest = ParseRepository(repository)
	assert.Equal(t, "", namespace)
	assert.Equal(t, "c", rest)
	// repository contains only one "/"
	repository = "b/c"
	namespace, rest = ParseRepository(repository)
	assert.Equal(t, "b", namespace)
	assert.Equal(t, "c", rest)
	// repository contains more than one "/"
	repository = "a/b/c"
	namespace, rest = ParseRepository(repository)
	assert.Equal(t, "a/b", namespace)
	assert.Equal(t, "c", rest)
}
