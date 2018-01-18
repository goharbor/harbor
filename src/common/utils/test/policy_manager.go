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

package test

import (
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/models"
)

type FakePolicyManager struct {
}

func (f *FakePolicyManager) GetPolicies(query models.QueryParameter) (*models.ReplicationPolicyQueryResult, error) {
	return &models.ReplicationPolicyQueryResult{}, nil
}

func (f *FakePolicyManager) GetPolicy(id int64) (models.ReplicationPolicy, error) {
	return models.ReplicationPolicy{
		ID: 1,
		Trigger: &models.Trigger{
			Kind: replication.TriggerKindManual,
		},
	}, nil
}
func (f *FakePolicyManager) CreatePolicy(policy models.ReplicationPolicy) (int64, error) {
	return 1, nil
}
func (f *FakePolicyManager) UpdatePolicy(models.ReplicationPolicy) error {
	return nil
}
func (f *FakePolicyManager) RemovePolicy(int64) error {
	return nil
}
