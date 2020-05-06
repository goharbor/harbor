package robot

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/stretchr/testify/assert"
)

func TestValid(t *testing.T) {

	rbacPolicy := &types.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*types.Policy{}
	policies = append(policies, rbacPolicy)

	rClaims := &Claim{
		TokenID:   1,
		ProjectID: 2,
		Access:    policies,
	}
	assert.Nil(t, rClaims.Valid())
}

func TestUnValidTokenID(t *testing.T) {

	rbacPolicy := &types.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*types.Policy{}
	policies = append(policies, rbacPolicy)

	rClaims := &Claim{
		TokenID:   -1,
		ProjectID: 2,
		Access:    policies,
	}
	assert.NotNil(t, rClaims.Valid())
}

func TestUnValidProjectID(t *testing.T) {

	rbacPolicy := &types.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*types.Policy{}
	policies = append(policies, rbacPolicy)

	rClaims := &Claim{
		TokenID:   1,
		ProjectID: -2,
		Access:    policies,
	}
	assert.NotNil(t, rClaims.Valid())
}

func TestUnValidPolicy(t *testing.T) {

	rClaims := &Claim{
		TokenID:   1,
		ProjectID: 2,
		Access:    nil,
	}
	assert.NotNil(t, rClaims.Valid())
}
