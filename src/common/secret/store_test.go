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

package secret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	store := NewStore(map[string]string{
		"secret1": "username1",
	})

	assert.False(t, store.IsValid("invalid_secret"))
	assert.True(t, store.IsValid("secret1"))
}

func TestGetUsername(t *testing.T) {
	store := NewStore(map[string]string{
		"secret1": "username1",
	})

	assert.Equal(t, "", store.GetUsername("invalid_secret"))
	assert.Equal(t, "username1", store.GetUsername("secret1"))
}
