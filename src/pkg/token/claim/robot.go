package claim

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/common/rbac"
)

// Robot implements the interface of jwt.Claims
type Robot struct {
	jwt.StandardClaims
	TokenID     int64          `json:"id"`
	ProjectID   int64          `json:"pid"`
	PolicyCheck bool           `json:"policy_check"`
	Access      []*rbac.Policy `json:"access"`
}

// Valid valid the claims "tokenID, projectID and access".
func (rc Robot) Valid() error {
	if rc.TokenID < 0 {
		return errors.New("Token id must an valid INT")
	}
	if rc.ProjectID < 0 {
		return errors.New("Project id must an valid INT")
	}
	if rc.Access == nil {
		return errors.New("The access info cannot be nil")
	}
	stdErr := rc.StandardClaims.Valid()
	if stdErr != nil {
		return stdErr
	}
	return nil
}
