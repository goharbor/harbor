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

package adapter

import (
	"testing"

	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakedFactory(*model.Registry) (Adapter, error) {
	return nil, nil
}

func TestRegisterFactory(t *testing.T) {
	// empty name
	assert.NotNil(t, RegisterFactory("", nil))
	// empty factory
	assert.NotNil(t, RegisterFactory("factory", nil))
	// pass
	assert.Nil(t, RegisterFactory("factory", fakedFactory))
	// already exists
	assert.NotNil(t, RegisterFactory("factory", fakedFactory))
}

func TestGetFactory(t *testing.T) {
	registry = map[model.RegistryType]Factory{}
	require.Nil(t, RegisterFactory("factory", fakedFactory))
	// doesn't exist
	_, err := GetFactory("another_factory")
	assert.NotNil(t, err)
	// pass
	_, err = GetFactory("factory")
	assert.Nil(t, err)
}
