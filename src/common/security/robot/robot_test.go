package robot

import (
	"testing"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/stretchr/testify/assert"
)

func TestGetPolicies(t *testing.T) {

	rbacPolicy := &rbac.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	robot := robot{
		username:  "test",
		namespace: rbac.NewProjectNamespace(1, false),
		policies:  policies,
	}

	assert.Equal(t, robot.GetUserName(), "test")
	assert.NotNil(t, robot.GetPolicies())
	assert.Nil(t, robot.GetRoles())
}

func TestNewRobot(t *testing.T) {
	policies := []*rbac.Policy{
		{Resource: "/project/1/repository", Action: "pull"},
		{Resource: "/project/library/repository", Action: "pull"},
		{Resource: "/project/library/repository", Action: "push"},
	}

	robot := NewRobot("test", rbac.NewProjectNamespace(1, false), policies)
	assert.Len(t, robot.GetPolicies(), 1)
}
