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

package model

import (
	"testing"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
)

func TestIsScheduledTrigger(t *testing.T) {
	assert := assert.New(t)

	// policy is disabled
	policy := &Policy{
		Enabled: false,
	}
	b := policy.IsScheduledTrigger()
	assert.False(b)

	// no trigger
	policy = &Policy{
		Enabled: true,
	}
	b = policy.IsScheduledTrigger()
	assert.False(b)

	// isn't scheduled trigger
	policy = &Policy{
		Trigger: &model.Trigger{
			Type: model.TriggerTypeEventBased,
		},
		Enabled: true,
	}
	b = policy.IsScheduledTrigger()
	assert.False(b)

	// scheduled trigger
	policy = &Policy{
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
		},
		Enabled: true,
	}
	b = policy.IsScheduledTrigger()
	assert.True(b)
}

func TestValidate(t *testing.T) {
	assert := assert.New(t)

	// empty name
	policy := &Policy{}
	err := policy.Validate()
	assert.True(errors.IsErr(err, errors.BadRequestCode))

	// empty source registry and destination registry
	policy = &Policy{
		Name: "policy01",
	}
	err = policy.Validate()
	assert.True(errors.IsErr(err, errors.BadRequestCode))

	// source registry and destination registry both not empty
	policy = &Policy{
		Name: "policy01",
		SrcRegistry: &model.Registry{
			ID: 1,
		},
		DestRegistry: &model.Registry{
			ID: 2,
		},
	}
	err = policy.Validate()
	assert.True(errors.IsErr(err, errors.BadRequestCode))

	// invalid filter
	policy = &Policy{
		Name: "policy01",
		SrcRegistry: &model.Registry{
			ID: 0,
		},
		DestRegistry: &model.Registry{
			ID: 1,
		},
		Filters: []*model.Filter{
			{
				Type: "invalid_type",
			},
		},
	}
	err = policy.Validate()
	assert.True(errors.IsErr(err, errors.BadRequestCode))

	// invalid filter
	policy = &Policy{
		Name: "policy01",
		SrcRegistry: &model.Registry{
			ID: 0,
		},
		DestRegistry: &model.Registry{
			ID: 1,
		},
		Filters: []*model.Filter{
			{
				Type:  model.FilterTypeResource,
				Value: "invalid_resource_type",
			},
		},
	}
	err = policy.Validate()
	assert.True(errors.IsErr(err, errors.BadRequestCode))

	// invalid trigger
	policy = &Policy{
		Name: "policy01",
		SrcRegistry: &model.Registry{
			ID: 0,
		},
		DestRegistry: &model.Registry{
			ID: 1,
		},
		Filters: []*model.Filter{
			{
				Type:  model.FilterTypeName,
				Value: "library",
			},
		},
		Trigger: &model.Trigger{
			Type: "invalid_type",
		},
	}
	err = policy.Validate()
	assert.True(errors.IsErr(err, errors.BadRequestCode))

	// invalid trigger
	policy = &Policy{
		Name: "policy01",
		SrcRegistry: &model.Registry{
			ID: 0,
		},
		DestRegistry: &model.Registry{
			ID: 1,
		},
		Filters: []*model.Filter{
			{
				Type:  model.FilterTypeName,
				Value: "library",
			},
		},
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
		},
	}
	err = policy.Validate()
	assert.True(errors.IsErr(err, errors.BadRequestCode))

	// invalid cron
	policy = &Policy{
		Name: "policy01",
		SrcRegistry: &model.Registry{
			ID: 0,
		},
		DestRegistry: &model.Registry{
			ID: 1,
		},
		Filters: []*model.Filter{
			{
				Type:  model.FilterTypeResource,
				Value: "image",
			},
			{
				Type:  model.FilterTypeName,
				Value: "library/**",
			},
		},
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
			Settings: &model.TriggerSettings{
				Cron: "* * *",
			},
		},
	}
	err = policy.Validate()
	assert.True(errors.IsErr(err, errors.BadRequestCode))

	// pass
	policy = &Policy{
		Name: "policy01",
		SrcRegistry: &model.Registry{
			ID: 0,
		},
		DestRegistry: &model.Registry{
			ID: 1,
		},
		Filters: []*model.Filter{
			{
				Type:  model.FilterTypeResource,
				Value: "image",
			},
			{
				Type:  model.FilterTypeName,
				Value: "library/**",
			},
		},
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
			Settings: &model.TriggerSettings{
				Cron: "* * * * * *",
			},
		},
	}
	err = policy.Validate()
	assert.Nil(err)
}
