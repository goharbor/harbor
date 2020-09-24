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

package robot

import (
	"testing"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/stretchr/testify/assert"
)

func TestGetPolicies(t *testing.T) {

	rbacPolicy := &types.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*types.Policy{}
	policies = append(policies, rbacPolicy)

	robot := robot{
		username:  "test",
		namespace: rbac.NewProjectNamespace(1),
		policies:  policies,
	}

	assert.Equal(t, robot.GetUserName(), "test")
	assert.NotNil(t, robot.GetPolicies())
	assert.Nil(t, robot.GetRoles())
}

func TestNewRobot(t *testing.T) {
	policies := []*types.Policy{
		{Resource: "/project/1/repository", Action: "push"},
		{Resource: "/project/1/repository", Action: "scanner-pull"},
		{Resource: "/project/library/repository", Action: "pull"},
		{Resource: "/project/library/repository", Action: "push"},
	}

	robot := NewRobot("test", rbac.NewProjectNamespace(1), policies)
	assert.Len(t, robot.GetPolicies(), 3)
}
