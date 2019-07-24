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

package hook

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/scheduler/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/require"
)

var h = &controller{
	manager: &htesting.FakeSchedulerManager{},
}

func TestUpdateStatus(t *testing.T) {
	// task not exist
	err := h.UpdateStatus(1, "running")
	require.NotNil(t, err)

	// pass
	h.manager.(*htesting.FakeSchedulerManager).Schedules = []*model.Schedule{
		{
			ID:     1,
			Status: "",
		},
	}
	err = h.UpdateStatus(1, "running")
	require.Nil(t, err)
}

func TestRun(t *testing.T) {
	// callback function not exist
	err := h.Run("not-exist", nil)
	require.NotNil(t, err)

	// pass
	err = scheduler.Register("callback", func(interface{}) error { return nil })
	require.Nil(t, err)
	err = h.Run("callback", nil)
	require.Nil(t, err)
}
