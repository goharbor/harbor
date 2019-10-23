package registry

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/docker/distribution/registry/auth"
	"github.com/docker/distribution/registry/auth/token"
)

// Claim implements the interface of jwt.Claims
type Claim struct {
	jwt.StandardClaims
	Access []*token.ResourceActions `json:"access"`
}

// Valid valid the standard claims
func (rc *Claim) Valid() error {
	return rc.StandardClaims.Valid()
}

// GetAccessSet ...
func (rc *Claim) GetAccessSet() AccessSet {
	accessSet := make(AccessSet, len(rc.Access))
	for _, resourceActions := range rc.Access {
		resource := auth.Resource{
			Type: resourceActions.Type,
			Name: resourceActions.Name,
		}
		set, exists := accessSet[resource]
		if !exists {
			set = newActionSet()
			accessSet[resource] = set
		}
		for _, action := range resourceActions.Actions {
			set.add(action)
		}
	}
	return accessSet
}
