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

package replication

import (
	repmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	replicationmodel "github.com/goharbor/harbor/src/pkg/replication/model"
	"github.com/goharbor/harbor/src/testing/mock"
)

func (r *replicationTestSuite) TestPolicyCount() {
	mock.OnAnything(r.repMgr, "Count").Return(int64(1), nil)
	count, err := r.ctl.PolicyCount(nil, nil)
	r.Require().Nil(err)
	r.Equal(int64(1), count)
	r.repMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestListPolicies() {
	mock.OnAnything(r.repMgr, "List").Return([]*replicationmodel.Policy{
		{
			ID:            1,
			SrcRegistryID: 1,
		},
	}, nil)
	mock.OnAnything(r.regMgr, "Get").Return(&model.Registry{
		ID: 1,
	}, nil)
	policies, err := r.ctl.ListPolicies(nil, nil)
	r.Require().Nil(err)
	r.Require().Len(policies, 1)
	r.Equal(int64(1), policies[0].ID)
	r.repMgr.AssertExpectations(r.T())
	r.regMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestGetPolicy() {
	mock.OnAnything(r.repMgr, "Get").Return(&replicationmodel.Policy{
		ID:            1,
		SrcRegistryID: 1,
	}, nil)
	mock.OnAnything(r.regMgr, "Get").Return(&model.Registry{
		ID: 1,
	}, nil)
	policy, err := r.ctl.GetPolicy(nil, 1)
	r.Require().Nil(err)
	r.Equal(int64(1), policy.ID)
	r.repMgr.AssertExpectations(r.T())
	r.regMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestCreatePolicy() {
	mock.OnAnything(r.repMgr, "Create").Return(int64(1), nil)
	mock.OnAnything(r.regMgr, "Get").Return(&model.Registry{
		ID: 1,
	}, nil)
	mock.OnAnything(r.scheduler, "Schedule").Return(int64(1), nil)
	id, err := r.ctl.CreatePolicy(nil, &repmodel.Policy{
		Name: "rule",
		SrcRegistry: &model.Registry{
			ID: 1,
		},
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
			Settings: &model.TriggerSettings{
				Cron: "0 * * * * *",
			},
		},
		Enabled: true,
	})
	r.Require().Nil(err)
	r.Equal(int64(1), id)
	r.repMgr.AssertExpectations(r.T())
	r.regMgr.AssertExpectations(r.T())
	r.scheduler.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestUpdatePolicy() {
	mock.OnAnything(r.regMgr, "Get").Return(&model.Registry{
		ID: 1,
	}, nil)
	mock.OnAnything(r.scheduler, "UnScheduleByVendor").Return(nil)
	mock.OnAnything(r.scheduler, "Schedule").Return(int64(1), nil)
	mock.OnAnything(r.repMgr, "Update").Return(nil)
	err := r.ctl.UpdatePolicy(nil, &repmodel.Policy{
		ID:   1,
		Name: "rule",
		SrcRegistry: &model.Registry{
			ID: 1,
		},
		Trigger: &model.Trigger{
			Type: model.TriggerTypeScheduled,
			Settings: &model.TriggerSettings{
				Cron: "0 * * * * *",
			},
		},
		Enabled: true,
	})
	r.Require().Nil(err)
	r.repMgr.AssertExpectations(r.T())
	r.regMgr.AssertExpectations(r.T())
	r.scheduler.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestDeletePolicy() {
	mock.OnAnything(r.execMgr, "DeleteByVendor").Return(nil)
	mock.OnAnything(r.scheduler, "UnScheduleByVendor").Return(nil)
	mock.OnAnything(r.repMgr, "Delete").Return(nil)
	err := r.ctl.DeletePolicy(nil, 1)
	r.Require().Nil(err)
	r.repMgr.AssertExpectations(r.T())
	r.execMgr.AssertExpectations(r.T())
	r.scheduler.AssertExpectations(r.T())
}
