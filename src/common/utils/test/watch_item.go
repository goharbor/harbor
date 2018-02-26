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
	"github.com/vmware/harbor/src/common/models"
)

// FakeWatchItemDAO is the fake implement for the dao.WatchItemDAO
type FakeWatchItemDAO struct {
	items []models.WatchItem
}

// Add ...
func (f *FakeWatchItemDAO) Add(item *models.WatchItem) (int64, error) {
	f.items = append(f.items, *item)
	return int64(len(f.items) + 1), nil
}

// DeleteByPolicyID : delete the WatchItem specified by policy ID
func (f *FakeWatchItemDAO) DeleteByPolicyID(policyID int64) error {
	for i, item := range f.items {
		if item.PolicyID == policyID {
			f.items = append(f.items[:i], f.items[i+1:]...)
			break
		}
	}
	return nil
}

// Get returns WatchItem list according to the namespace and operation
func (f *FakeWatchItemDAO) Get(namespace, operation string) ([]models.WatchItem, error) {
	items := []models.WatchItem{}
	for _, item := range f.items {
		if item.Namespace != namespace {
			continue
		}

		if operation == "push" {
			if item.OnPush {
				items = append(items, item)
			}
		}

		if operation == "delete" {
			if item.OnDeletion {
				items = append(items, item)
			}
		}
	}

	return items, nil
}
