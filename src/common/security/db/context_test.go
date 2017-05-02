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

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/models"
)

type fakePMS struct {
	public string
}

func (f *fakePMS) IsPublic(projectIDOrName interface{}) bool {
	return f.public == projectIDOrName.(string)
}
func (f *fakePMS) HasReadPerm(username string, projectIDOrName interface{},
	token ...string) bool {
	return true
}
func (f *fakePMS) HasWritePerm(username string, projectIDOrName interface{},
	token ...string) bool {
	return true
}
func (f *fakePMS) HasAllPerm(username string, projectIDOrName interface{},
	token ...string) bool {
	return true
}

func TestIsAuthenticated(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil)
	assert.False(t, ctx.IsAuthenticated())

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, nil)
	assert.True(t, ctx.IsAuthenticated())
}

func TestGetUsername(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil)
	assert.Equal(t, "", ctx.GetUsername())

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, nil)
	assert.Equal(t, "test", ctx.GetUsername())
}

func TestIsSysAdmin(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, non admin
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, admin
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		HasAdminRole: 1,
	}, nil)
	assert.True(t, ctx.IsSysAdmin())
}

func TestHasReadPerm(t *testing.T) {
	pms := &fakePMS{
		public: "public_project",
	}

	// public project, unauthenticated
	ctx := NewSecurityContext(nil, pms)
	assert.True(t, ctx.HasReadPerm("public_project"))

	// private project, unauthenticated
	ctx = NewSecurityContext(nil, pms)
	assert.False(t, ctx.HasReadPerm("private_project"))

	// private project, authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pms)
	assert.True(t, ctx.HasReadPerm("private_project"))
}

func TestHasWritePerm(t *testing.T) {
	pms := &fakePMS{}

	// unauthenticated
	ctx := NewSecurityContext(nil, pms)
	assert.False(t, ctx.HasWritePerm("project"))

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pms)
	assert.True(t, ctx.HasWritePerm("project"))
}

func TestHasAllPerm(t *testing.T) {
	pms := &fakePMS{}

	// unauthenticated
	ctx := NewSecurityContext(nil, pms)
	assert.False(t, ctx.HasAllPerm("project"))

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pms)
	assert.True(t, ctx.HasAllPerm("project"))
}
