// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package trigger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/test"
	"github.com/vmware/harbor/src/replication"
)

func TestKindOfImmediateTrigger(t *testing.T) {
	trigger := NewImmediateTrigger(ImmediateParam{})
	assert.Equal(t, replication.TriggerKindImmediate, trigger.Kind())
}

func TestSetupAndUnsetOfImmediateTrigger(t *testing.T) {
	dao.DefaultDatabaseWatchItemDAO = &test.FakeWatchItemDAO{}

	param := ImmediateParam{}
	param.PolicyID = 1
	param.OnDeletion = true
	param.Namespaces = []string{"library"}
	trigger := NewImmediateTrigger(param)

	err := trigger.Setup()
	require.Nil(t, err)

	items, err := DefaultWatchList.Get("library", "push")
	require.Nil(t, err)
	assert.Equal(t, 1, len(items))

	items, err = DefaultWatchList.Get("library", "delete")
	require.Nil(t, err)
	assert.Equal(t, 1, len(items))

	err = trigger.Unset()
	require.Nil(t, err)
	items, err = DefaultWatchList.Get("library", "delete")
	require.Nil(t, err)
	assert.Equal(t, 0, len(items))
}
