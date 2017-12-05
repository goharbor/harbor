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

type FakeReplicatoinController struct {
	FakePolicyManager
}

func (f *FakeReplicatoinController) Init() error {
	return nil
}
func (f *FakeReplicatoinController) Replicate(policyID int64, metadata ...map[string]interface{}) error {
	return nil
}
