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

package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRegistryClient(t *testing.T) {
	// failure case with invalid address
	t.Setenv("_REDIS_URL_REG", "invalid-address")
	client, err := GetRegistryClient()
	assert.Error(t, err)
	assert.Nil(t, client)

	// normal case with valid address
	t.Setenv("_REDIS_URL_REG", "redis://localhost:6379/1")
	client, err = GetRegistryClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// multiple calls should return the same client
	for i := 0; i < 10; i++ {
		newClient, err := GetRegistryClient()
		assert.NoError(t, err)
		assert.Equal(t, client, newClient)
	}
}

func TestGetHarborClient(t *testing.T) {
	// failure case with invalid address
	t.Setenv("_REDIS_URL_HARBOR", "invalid-address")
	client, err := GetHarborClient()
	assert.Error(t, err)
	assert.Nil(t, client)

	// normal case with valid address
	t.Setenv("_REDIS_URL_HARBOR", "redis://localhost:6379/0")
	client, err = GetHarborClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// multiple calls should return the same client
	for i := 0; i < 10; i++ {
		newClient, err := GetHarborClient()
		assert.NoError(t, err)
		assert.Equal(t, client, newClient)
	}
}
