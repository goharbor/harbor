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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/scheduler"
	"github.com/vmware/harbor/src/replication"
)

func TestAssembleName(t *testing.T) {
	assert.Equal(t, "replication_policy_1", assembleName(1))
}

func TestKindOfScheduleTrigger(t *testing.T) {
	trigger := NewScheduleTrigger(ScheduleParam{})
	assert.Equal(t, replication.TriggerKindSchedule, trigger.Kind())
}

func TestSetupAndUnSetOfScheduleTrigger(t *testing.T) {
	// invalid schedule param
	trigger := NewScheduleTrigger(ScheduleParam{})
	assert.NotNil(t, trigger.Setup())

	// valid schedule param
	var policyID int64 = 1
	trigger = NewScheduleTrigger(ScheduleParam{
		BasicParam: BasicParam{
			PolicyID: policyID,
		},
		Type:    replication.TriggerScheduleWeekly,
		Weekday: (int8(time.Now().Weekday()) + 1) % 7,
		Offtime: 0,
	})

	count := scheduler.DefaultScheduler.PolicyCount()
	require.Nil(t, scheduler.DefaultScheduler.GetPolicy(assembleName(policyID)))

	require.Nil(t, trigger.Setup())

	assert.Equal(t, count+1, scheduler.DefaultScheduler.PolicyCount())
	assert.NotNil(t, scheduler.DefaultScheduler.GetPolicy(assembleName(policyID)))

	require.Nil(t, trigger.Unset())
	assert.Equal(t, count, scheduler.DefaultScheduler.PolicyCount())
	assert.Nil(t, scheduler.DefaultScheduler.GetPolicy(assembleName(policyID)))
}
