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

package secret

import (
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/secret"
	"github.com/stretchr/testify/assert"
)

func TestIsAuthenticated(t *testing.T) {
	// secret store is null
	context := NewSecurityContext("", nil)
	isAuthenticated := context.IsAuthenticated()
	assert.False(t, isAuthenticated)

	// invalid secret
	context = NewSecurityContext("invalid_secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))
	isAuthenticated = context.IsAuthenticated()
	assert.False(t, isAuthenticated)

	// valid secret
	context = NewSecurityContext("secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))
	isAuthenticated = context.IsAuthenticated()
	assert.True(t, isAuthenticated)
}

func TestGetUsername(t *testing.T) {
	// secret store is null
	context := NewSecurityContext("", nil)
	username := context.GetUsername()
	assert.Equal(t, "", username)

	// invalid secret
	context = NewSecurityContext("invalid_secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))
	username = context.GetUsername()
	assert.Equal(t, "", username)

	// valid secret
	context = NewSecurityContext("secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))
	username = context.GetUsername()
	assert.Equal(t, "username", username)
}

func TestIsSysAdmin(t *testing.T) {
	context := NewSecurityContext("secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))
	isSysAdmin := context.IsSysAdmin()
	assert.False(t, isSysAdmin)
}

func TestIsSolutionUser(t *testing.T) {
	// invalid secret
	context := NewSecurityContext("invalid_secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))
	isSolutionUser := context.IsSolutionUser()
	assert.False(t, isSolutionUser)

	// valid secret
	context = NewSecurityContext("secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))
	isSolutionUser = context.IsSolutionUser()
	assert.True(t, isSolutionUser)
}

func TestHasReadPerm(t *testing.T) {
	readAction := rbac.Action("pull")
	resource := rbac.Resource("/project/project_name/repository")
	// secret store is null
	context := NewSecurityContext("", nil)
	hasReadPerm := context.Can(readAction, resource)
	assert.False(t, hasReadPerm)

	// invalid secret
	context = NewSecurityContext("invalid_secret",
		secret.NewStore(map[string]string{
			"jobservice_secret": secret.JobserviceUser,
		}))
	hasReadPerm = context.Can(readAction, resource)
	assert.False(t, hasReadPerm)

	// valid secret, project name
	context = NewSecurityContext("jobservice_secret",
		secret.NewStore(map[string]string{
			"jobservice_secret": secret.JobserviceUser,
		}))
	hasReadPerm = context.Can(readAction, resource)
	assert.True(t, hasReadPerm)

	// valid secret, project ID
	resource = rbac.Resource("/project/1/repository")
	hasReadPerm = context.Can(readAction, resource)
	assert.True(t, hasReadPerm)
}

func TestHasWritePerm(t *testing.T) {
	context := NewSecurityContext("secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))

	writeAction := rbac.Action("push")

	// project name
	resource := rbac.Resource("/project/project_name/repository")
	hasWritePerm := context.Can(writeAction, resource)
	assert.False(t, hasWritePerm)

	// project ID
	resource = rbac.Resource("/project/1/repository")
	hasWritePerm = context.Can(writeAction, resource)
	assert.False(t, hasWritePerm)
}

func TestHasAllPerm(t *testing.T) {
	context := NewSecurityContext("secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))

	allAction := rbac.Action("push+pull")

	// project name
	resource := rbac.Resource("/project/project_name/repository")
	hasAllPerm := context.Can(allAction, resource)
	assert.False(t, hasAllPerm)

	// project ID
	resource = rbac.Resource("/project/1/repository")
	hasAllPerm = context.Can(allAction, resource)
	assert.False(t, hasAllPerm)
}

func TestGetMyProjects(t *testing.T) {
	context := NewSecurityContext("secret",
		secret.NewStore(map[string]string{
			"secret": "username",
		}))

	_, err := context.GetMyProjects()
	assert.NotNil(t, err)
}

func TestGetProjectRoles(t *testing.T) {
	// invalid secret
	context := NewSecurityContext("invalid_secret",
		secret.NewStore(map[string]string{
			"jobservice_secret": secret.JobserviceUser,
		}))

	roles := context.GetProjectRoles("any_project")
	assert.Equal(t, 0, len(roles))

	// valid secret
	context = NewSecurityContext("jobservice_secret",
		secret.NewStore(map[string]string{
			"jobservice_secret": secret.JobserviceUser,
		}))

	roles = context.GetProjectRoles("any_project")
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, common.RoleGuest, roles[0])
}
