package robot

import (
	"errors"

	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/golang-jwt/jwt/v4"
)

// Claim implements the interface of jwt.Claims
type Claim struct {
	jwt.StandardClaims
	TokenID   int64           `json:"id"`
	ProjectID int64           `json:"pid"`
	Access    []*types.Policy `json:"access"`
}

// Valid valid the claims "tokenID, projectID and access".
func (rc Claim) Valid() error {
	if rc.TokenID < 0 {
		return errors.New("token id must an valid INT")
	}
	if rc.ProjectID < 0 {
		return errors.New("project id must an valid INT")
	}
	if rc.Access == nil {
		return errors.New("the access info cannot be nil")
	}
	stdErr := rc.StandardClaims.Valid()
	if stdErr != nil {
		return stdErr
	}
	return nil
}
