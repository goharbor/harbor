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
package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/rbac"
	roleModel "github.com/goharbor/harbor/src/pkg/role/model"
)

func TestRoleEventsResolveToAuditLog(t *testing.T) {
	now := time.Now()
	role := &roleModel.Role{}
	role.Name = "auditor"

	create := &CreateRoleEvent{Role: role, Operator: "admin", OccurAt: now}
	cl, err := create.ResolveToAuditLog()
	assert.NoError(t, err)
	assert.Equal(t, rbac.ActionCreate.String(), cl.Operation)
	assert.Equal(t, ResourceTypeRole, cl.ResourceType)
	assert.Equal(t, "auditor", cl.Resource)
	assert.Equal(t, "admin", cl.Username)
	assert.True(t, cl.IsSuccessful)
	assert.Contains(t, create.String(), "auditor")

	update := &UpdateRoleEvent{Role: role, Operator: "admin", OccurAt: now}
	ul, err := update.ResolveToAuditLog()
	assert.NoError(t, err)
	assert.Equal(t, rbac.ActionUpdate.String(), ul.Operation)
	assert.Equal(t, ResourceTypeRole, ul.ResourceType)
	assert.Contains(t, update.String(), "auditor")

	del := &DeleteRoleEvent{Role: role, Operator: "admin", OccurAt: now}
	dl, err := del.ResolveToAuditLog()
	assert.NoError(t, err)
	assert.Equal(t, rbac.ActionDelete.String(), dl.Operation)
	assert.Equal(t, ResourceTypeRole, dl.ResourceType)
	assert.Contains(t, del.String(), "auditor")
}
