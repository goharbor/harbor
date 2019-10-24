package token

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValid(t *testing.T) {

	rbacPolicy := &rbac.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	rClaims := &RobotClaims{
		TokenID:   1,
		ProjectID: 2,
		Access:    policies,
	}
	assert.Nil(t, rClaims.Valid())
}

func TestUnValidTokenID(t *testing.T) {

	rbacPolicy := &rbac.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	rClaims := &RobotClaims{
		TokenID:   -1,
		ProjectID: 2,
		Access:    policies,
	}
	assert.NotNil(t, rClaims.Valid())
}

func TestUnValidProjectID(t *testing.T) {

	rbacPolicy := &rbac.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	rClaims := &RobotClaims{
		TokenID:   1,
		ProjectID: -2,
		Access:    policies,
	}
	assert.NotNil(t, rClaims.Valid())
}

func TestUnValidPolicy(t *testing.T) {

	rClaims := &RobotClaims{
		TokenID:   1,
		ProjectID: 2,
		Access:    nil,
	}
	assert.NotNil(t, rClaims.Valid())
}
