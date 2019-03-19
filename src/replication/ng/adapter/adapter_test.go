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
	// empty type
	assert.NotNil(t, RegisterFactory(&Info{}, nil))
	// empty supportted resource type
	assert.NotNil(t, RegisterFactory(
		&Info{
			Type: "harbor",
		}, nil))
	// empty trigger
	assert.NotNil(t, RegisterFactory(
		&Info{
			Type:                   "harbor",
			SupportedResourceTypes: []model.ResourceType{"image"},
		}, nil))
	// empty factory
	assert.NotNil(t, RegisterFactory(
		&Info{
			Type:                   "harbor",
			SupportedResourceTypes: []model.ResourceType{"image"},
			SupportedTriggers:      []model.TriggerType{"mannual"},
		}, nil))
	// pass
	assert.Nil(t, RegisterFactory(
		&Info{
			Type:                   "harbor",
			SupportedResourceTypes: []model.ResourceType{"image"},
			SupportedTriggers:      []model.TriggerType{"mannual"},
		}, fakedFactory))
	// already exists
	assert.NotNil(t, RegisterFactory(
		&Info{
			Type:                   "harbor",
			SupportedResourceTypes: []model.ResourceType{"image"},
			SupportedTriggers:      []model.TriggerType{"mannual"},
		}, fakedFactory))
}

func TestGetFactory(t *testing.T) {
	registry = []*item{}
	require.Nil(t, RegisterFactory(
		&Info{
			Type:                   "harbor",
			SupportedResourceTypes: []model.ResourceType{"image"},
			SupportedTriggers:      []model.TriggerType{"mannual"},
		}, fakedFactory))
	// doesn't exist
	_, err := GetFactory("gcr")
	assert.NotNil(t, err)
	// pass
	_, err = GetFactory("harbor")
	assert.Nil(t, err)
}

func TestListAdapterInfos(t *testing.T) {
	registry = []*item{}
	// not register, got nothing
	infos := ListAdapterInfos()
	assert.Equal(t, 0, len(infos))

	// register one factory
	require.Nil(t, RegisterFactory(
		&Info{
			Type:                   "harbor",
			SupportedResourceTypes: []model.ResourceType{"image"},
			SupportedTriggers:      []model.TriggerType{"mannual"},
		}, fakedFactory))

	infos = ListAdapterInfos()
	require.Equal(t, 1, len(infos))
	assert.Equal(t, "harbor", string(infos[0].Type))
}

func TestGetAdapterInfo(t *testing.T) {
	registry = []*item{}
	require.Nil(t, RegisterFactory(
		&Info{
			Type:                   "harbor",
			SupportedResourceTypes: []model.ResourceType{"image"},
			SupportedTriggers:      []model.TriggerType{"mannual"},
		}, fakedFactory))

	// doesn't exist
	info := GetAdapterInfo("gcr")
	assert.Nil(t, info)

	// exist
	info = GetAdapterInfo("harbor")
	require.NotNil(t, info)
	assert.Equal(t, "harbor", string(info.Type))
}
