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

package controller

import (
	"testing"

	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedPolicyController struct {
	policy *model.Policy
}

func (f *fakedPolicyController) Create(*model.Policy) (int64, error) {
	return 0, nil
}
func (f *fakedPolicyController) List(...*model.PolicyQuery) (int64, []*model.Policy, error) {
	return 0, nil, nil
}
func (f *fakedPolicyController) Get(id int64) (*model.Policy, error) {
	return f.policy, nil
}
func (f *fakedPolicyController) GetByName(name string) (*model.Policy, error) {
	return nil, nil
}
func (f *fakedPolicyController) Update(*model.Policy) error {
	return nil
}
func (f *fakedPolicyController) Remove(int64) error {
	return nil
}

type fakedScheduler struct {
	scheduled   bool
	unscheduled bool
}

func (f *fakedScheduler) Schedule(policyID int64, cron string) error {
	f.scheduled = true
	return nil
}
func (f *fakedScheduler) Unschedule(policyID int64) error {
	f.unscheduled = true
	return nil
}

func TestIsScheduledTrigger(t *testing.T) {
	cases := []struct {
		policy   *model.Policy
		expected bool
	}{
		// policy is nil
		{
			policy:   nil,
			expected: false,
		},
		// policy is disabled
		{
			policy: &model.Policy{
				Enabled: false,
			},
			expected: false,
		},
		// trigger is nil
		{
			policy: &model.Policy{
				Enabled: true,
			},
			expected: false,
		},
		// trigger type isn't scheduled
		{
			policy: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeManual,
				},
			},
			expected: false,
		},
		// trigger type is scheduled
		{
			policy: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeScheduled,
				},
			},
			expected: true,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expected, isScheduledTrigger(c.policy))
	}
}

func TestIsScheduleTriggerChanged(t *testing.T) {
	cases := []struct {
		origin   *model.Policy
		current  *model.Policy
		expected bool
	}{
		// both triggers are not scheduled
		{
			origin: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeManual,
				},
			},
			current: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeManual,
				},
			},
			expected: false,
		},
		// both triggers are scheduled and the crons are not same
		{
			origin: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeScheduled,
					Settings: &model.TriggerSettings{
						Cron: "03 05 * * *",
					},
				},
			},
			current: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeScheduled,
					Settings: &model.TriggerSettings{
						Cron: "03 * * * *",
					},
				},
			},
			expected: true,
		},
		// both triggers are scheduled and the crons are same
		{
			origin: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeScheduled,
					Settings: &model.TriggerSettings{
						Cron: "03 05 * * *",
					},
				},
			},
			current: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeScheduled,
					Settings: &model.TriggerSettings{
						Cron: "03 05 * * *",
					},
				},
			},
			expected: false,
		},
		// one trigger is scheduled but the other one isn't
		{
			origin: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeScheduled,
					Settings: &model.TriggerSettings{
						Cron: "03 05 * * *",
					},
				},
			},
			current: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeManual,
				},
			},
			expected: true,
		},
		// one trigger is scheduled but disabled and
		// the other one is scheduled but enabled
		{
			origin: &model.Policy{
				Enabled: false,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeScheduled,
					Settings: &model.TriggerSettings{
						Cron: "03 05 * * *",
					},
				},
			},
			current: &model.Policy{
				Enabled: true,
				Trigger: &model.Trigger{
					Type: model.TriggerTypeScheduled,
					Settings: &model.TriggerSettings{
						Cron: "03 05 * * *",
					},
				},
			},
			expected: true,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expected, isScheduleTriggerChanged(c.origin, c.current))
	}
}

func TestCreate(t *testing.T) {
	scheduler := &fakedScheduler{}
	ctl := &controller{
		scheduler: scheduler,
	}
	ctl.Controller = &fakedPolicyController{}

	// not scheduled trigger
	_, err := ctl.Create(&model.Policy{})
	require.Nil(t, err)
	assert.False(t, scheduler.scheduled)

	// scheduled trigger
	_, err = ctl.Create(&model.Policy{
		Enabled: true,
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
			Settings: &model.TriggerSettings{
				Cron: "03 05 * * *",
			},
		},
	})
	require.Nil(t, err)
	assert.True(t, scheduler.scheduled)
}

func TestUpdate(t *testing.T) {
	scheduler := &fakedScheduler{}
	c := &fakedPolicyController{}
	ctl := &controller{
		scheduler: scheduler,
	}
	ctl.Controller = c

	var origin, current *model.Policy
	// origin policy is nil
	current = &model.Policy{
		ID:      1,
		Enabled: true,
	}
	err := ctl.Update(current)
	assert.NotNil(t, err)

	// the trigger doesn't change
	origin = &model.Policy{
		ID:      1,
		Enabled: true,
	}
	c.policy = origin
	current = origin
	err = ctl.Update(current)
	require.Nil(t, err)
	assert.False(t, scheduler.scheduled)
	assert.False(t, scheduler.unscheduled)

	// the trigger changed
	origin = &model.Policy{
		ID:      1,
		Enabled: true,
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
			Settings: &model.TriggerSettings{
				Cron: "03 05 * * *",
			},
		},
	}
	c.policy = origin
	current = &model.Policy{
		Enabled: true,
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
			Settings: &model.TriggerSettings{
				Cron: "03 * * * *",
			},
		},
	}
	err = ctl.Update(current)
	require.Nil(t, err)
	assert.True(t, scheduler.unscheduled)
	assert.True(t, scheduler.scheduled)
}

func TestRemove(t *testing.T) {
	scheduler := &fakedScheduler{}
	c := &fakedPolicyController{}
	ctl := &controller{
		scheduler: scheduler,
	}
	ctl.Controller = c

	// policy is nil
	err := ctl.Remove(1)
	assert.NotNil(t, err)

	// the trigger type isn't scheduled
	policy := &model.Policy{
		ID:      1,
		Enabled: true,
		Trigger: &model.Trigger{
			Type: model.TriggerTypeManual,
		},
	}
	c.policy = policy
	err = ctl.Remove(1)
	require.Nil(t, err)
	assert.False(t, scheduler.unscheduled)

	// the trigger type is scheduled
	policy = &model.Policy{
		ID:      1,
		Enabled: true,
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
			Settings: &model.TriggerSettings{
				Cron: "03 05 * * *",
			},
		},
	}
	c.policy = policy
	err = ctl.Remove(1)
	require.Nil(t, err)
	assert.True(t, scheduler.unscheduled)
}
