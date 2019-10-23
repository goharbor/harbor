package registry

import (
	"github.com/docker/distribution/registry/auth"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValid(t *testing.T) {
	access := &token.ResourceActions{
		Type:    "type",
		Name:    "repository",
		Actions: []string{"pull", "push"},
	}
	accesses := []*token.ResourceActions{}
	accesses = append(accesses, access)
	rClaims := &Claim{
		Access: accesses,
	}
	assert.Nil(t, rClaims.Valid())
}

func TestGetAccessSet(t *testing.T) {
	access := &token.ResourceActions{
		Type:    "repository",
		Name:    "hello-world",
		Actions: []string{"pull", "push", "scanner-pull"},
	}
	accesses := []*token.ResourceActions{}
	accesses = append(accesses, access)
	rClaims := &Claim{
		Access: accesses,
	}

	auth1 := auth.Access{
		Resource: auth.Resource{
			Type: "repository",
			Name: "hello-world",
		},
		Action: rbac.ActionScannerPull.String(),
	}
	auth2 := auth.Access{
		Resource: auth.Resource{
			Type: "repository",
			Name: "busubox",
		},
		Action: rbac.ActionScannerPull.String(),
	}
	set := rClaims.GetAccessSet()
	assert.True(t, set.Contains(auth1))
	assert.False(t, set.Contains(auth2))
}
