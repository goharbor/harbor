package oidc

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
)
import "errors"

// This is for testing only
type fakeVerifier struct {
	secret string
}

func (fv *fakeVerifier) VerifySecret(ctx context.Context, name string, secret string) (*models.User, error) {
	if secret != fv.secret {
		return nil, verifyError(errors.New("mismatch"))
	}
	return &models.User{UserID: 1, Username: name, Email: fmt.Sprintf("%s@test.local", name)}, nil
}

// SetHardcodeVerifierForTest overwrite the default secret manager for testing.
// Be reminded this is for testing only.
func SetHardcodeVerifierForTest(s string) {
	m = &fakeVerifier{s}
}
