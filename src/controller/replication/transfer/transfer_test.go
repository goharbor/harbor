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

package transfer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fakedFactory Factory = func(Logger, StopFunc) (Transfer, error) {
	return nil, nil
}

func TestRegisterFactory(t *testing.T) {
	// empty name
	err := RegisterFactory("", fakedFactory)
	require.NotNil(t, err)
	// nil factory
	err = RegisterFactory("faked_factory", nil)
	require.NotNil(t, err)
	// pass
	err = RegisterFactory("faked_factory", fakedFactory)
	require.Nil(t, err)
	// already exist
	err = RegisterFactory("faked_factory", fakedFactory)
	require.NotNil(t, err)
}

func TestGetFactory(t *testing.T) {
	registry = map[string]Factory{}
	err := RegisterFactory("faked_factory", fakedFactory)
	require.Nil(t, err)
	// try to get the factory that doesn't exist
	_, err = GetFactory("not_exist_factory")
	assert.NotNil(t, err)
	// pass
	_, err = GetFactory("faked_factory")
	require.Nil(t, err)
}
