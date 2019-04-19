package oidc

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
)
import "errors"

// This is for testing only
type fakeVerifier struct {
	secret string
}

func (fv *fakeVerifier) VerifySecret(ctx context.Context, userID int, secret string) error {
	if secret != fv.secret {
		return verifyError(errors.New("mismatch"))
	}
	return nil
}

func (fv *fakeVerifier) VerifyToken(ctx context.Context, u *models.OIDCUser) error {
	return nil
}

// SetHardcodeVerifierForTest overwrite the default secret manager for testing.
// Be reminded this is for testing only.
func SetHardcodeVerifierForTest(s string) {
	m = &fakeVerifier{s}
}
