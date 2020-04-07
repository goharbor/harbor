package v2token

import (
	"testing"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestAll(t *testing.T) {
	mgr := &project.FakeManager{}
	mgr.On("Get", int64(1)).Return(&models.Project{ProjectID: 1, Name: "library"}, nil)
	mgr.On("Get", int64(2)).Return(&models.Project{ProjectID: 2, Name: "test"}, nil)
	mgr.On("Get", int64(3)).Return(&models.Project{ProjectID: 3, Name: "development"}, nil)

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
	tsc.pm = mgr

	cases := []struct {
		resource types.Resource
		action   types.Action
		expect   bool
	}{
		{
			resource: rbac.NewProjectNamespace(1).Resource(rbac.ResourceRepository),
			action:   rbac.ActionPush,
			expect:   true,
		},
		{
			resource: rbac.NewProjectNamespace(1).Resource(rbac.ResourceRepository),
			action:   rbac.ActionScannerPull,
			expect:   true,
		},
		{
			resource: rbac.NewProjectNamespace(2).Resource(rbac.ResourceRepository),
			action:   rbac.ActionPush,
			expect:   true,
		},
		{
			resource: rbac.NewProjectNamespace(2).Resource(rbac.ResourceRepository),
			action:   rbac.ActionScannerPull,
			expect:   false,
		},
		{
			resource: rbac.NewProjectNamespace(3).Resource(rbac.ResourceRepository),
			action:   rbac.ActionPush,
			expect:   false,
		},
		{
			resource: rbac.NewProjectNamespace(2).Resource(rbac.ResourceArtifact),
			action:   rbac.ActionPush,
			expect:   false,
		},
		{
			resource: rbac.NewProjectNamespace(1).Resource(rbac.ResourceRepository),
			action:   rbac.ActionCreate,
			expect:   false,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.expect, sc.Can(c.action, c.resource))
	}
}
