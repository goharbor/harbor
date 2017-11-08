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

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalAndUnmarshal(t *testing.T) {
	trigger := &RepTrigger{
		Type:   "schedule",
		Params: map[string]interface{}{"date": "2:00"},
	}
	filters := []*RepFilter{
		&RepFilter{
			Type:  "repository",
			Value: "library/ubuntu*",
		},
	}
	policy := &RepPolicy{
		Trigger: trigger,
		Filters: filters,
	}

	err := policy.Marshal()
	require.Nil(t, err)

	policy.Trigger = nil
	policy.Filters = nil
	err = policy.Unmarshal()
	require.Nil(t, err)

	assert.EqualValues(t, filters, policy.Filters)
	assert.EqualValues(t, trigger, policy.Trigger)
}
