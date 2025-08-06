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

package v2token

import (
	"context"
	"testing"

	"github.com/distribution/distribution/v3/registry/auth/token"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/rbac"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/testing/controller/project"
)

func TestAll(t *testing.T) {
	ctx := context.TODO()

	ctl := &project.Controller{}
	ctl.On("Get", ctx, int64(1)).Return(&models.Project{ProjectID: 1, Name: "library"}, nil)
	ctl.On("Get", ctx, int64(2)).Return(&models.Project{ProjectID: 2, Name: "test"}, nil)
	ctl.On("Get", ctx, int64(3)).Return(&models.Project{ProjectID: 3, Name: "development"}, nil)

	access := []*token.ResourceActions{
		{
			Type: "repository",
			Name: "library/ubuntu",
			Actions: []string{
				"pull",
				"push",
				"scanner-pull",
			},
		},
		{
			Type: "repository",
			Name: "test/golang",
			Actions: []string{
				"pull",
				"*",
			},
		},
		{
			Type: "cnab",
			Name: "development/cnab",
			Actions: []string{
				"pull",
				"push",
			},
		},
	}
	sc := New(context.Background(), "jack", access)
	tsc := sc.(*tokenSecurityCtx)
	tsc.ctl = ctl

	cases := []struct {
		resource types.Resource
		action   types.Action
		expect   bool
	}{
		{
			resource: rbac_project.NewNamespace(1).Resource(rbac.ResourceRepository),
			action:   rbac.ActionPush,
			expect:   true,
		},
		{
			resource: rbac_project.NewNamespace(1).Resource(rbac.ResourceRepository),
			action:   rbac.ActionScannerPull,
			expect:   true,
		},
		{
			resource: rbac_project.NewNamespace(2).Resource(rbac.ResourceRepository),
			action:   rbac.ActionPush,
			expect:   true,
		},
		{
			resource: rbac_project.NewNamespace(2).Resource(rbac.ResourceRepository),
			action:   rbac.ActionDelete,
			expect:   true,
		},
		{
			resource: rbac_project.NewNamespace(2).Resource(rbac.ResourceRepository),
			action:   rbac.ActionScannerPull,
			expect:   false,
		},
		{
			resource: rbac_project.NewNamespace(3).Resource(rbac.ResourceRepository),
			action:   rbac.ActionPush,
			expect:   false,
		},
		{
			resource: rbac_project.NewNamespace(2).Resource(rbac.ResourceArtifact),
			action:   rbac.ActionPush,
			expect:   false,
		},
		{
			resource: rbac_project.NewNamespace(1).Resource(rbac.ResourceRepository),
			action:   rbac.ActionCreate,
			expect:   false,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.expect, sc.Can(ctx, c.action, c.resource))
	}
}
