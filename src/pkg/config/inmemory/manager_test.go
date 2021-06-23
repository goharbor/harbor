//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package inmemory

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewInMemoryManager(t *testing.T) {
	ctx := context.Background()
	inMemoryManager := NewInMemoryManager()
	inMemoryManager.UpdateConfig(ctx, map[string]interface{}{
		"ldap_url":         "ldaps://ldap.vmware.com",
		"ldap_timeout":     5,
		"ldap_verify_cert": true,
	})
	assert.Equal(t, "ldaps://ldap.vmware.com", inMemoryManager.Get(ctx, "ldap_url").GetString())
	assert.Equal(t, 5, inMemoryManager.Get(ctx, "ldap_timeout").GetInt())
	assert.Equal(t, true, inMemoryManager.Get(ctx, "ldap_verify_cert").GetBool())
}
