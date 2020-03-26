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

package security

import (
	"context"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// Context abstracts the operations related with authN and authZ
type Context interface {
	// Name returns the name of the security context
	Name() string
	// IsAuthenticated returns whether the context has been authenticated or not
	IsAuthenticated() bool
	// GetUsername returns the username of user related to the context
	GetUsername() string
	// IsSysAdmin returns whether the user is system admin
	IsSysAdmin() bool
	// IsSolutionUser returns whether the user is solution user
	IsSolutionUser() bool
	// Get current user's all project
	GetMyProjects() ([]*models.Project, error)
	// Get user's role in provided project
	GetProjectRoles(projectIDOrName interface{}) []int
	// Can returns whether the user can do action on resource
	Can(action types.Action, resource types.Resource) bool
}

type securityKey struct{}

// NewContext returns context with security context
func NewContext(ctx context.Context, security Context) context.Context {
	return context.WithValue(ctx, securityKey{}, security)
}

// FromContext returns security context from the context
func FromContext(ctx context.Context) (Context, bool) {
	c, ok := ctx.Value(securityKey{}).(Context)
	return c, ok
}
