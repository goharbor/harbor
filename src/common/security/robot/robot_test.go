package robot

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPolicies(t *testing.T) {

	rbacPolicy := &rbac.Policy{
		Resource: "/project/library/repository",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	robot := robot{
		username:  "test",
		namespace: rbac.NewProjectNamespace("library", false),
		policy:    policies,
	}

	assert.Equal(t, robot.GetUserName(), "test")
	assert.NotNil(t, robot.GetPolicies())
	assert.Nil(t, robot.GetRoles())
}
