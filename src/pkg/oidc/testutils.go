package oidc

import (
	"context"
	"fmt"
	"strconv"
)
import "errors"

// This is for testing only
type fakeVerifier struct {
	secret string
}

func (fv *fakeVerifier) VerifySecret(ctx context.Context, name string, secret string) (*UserInfo, error) {
	if secret != fv.secret {
		return nil, verifyError(errors.New("mismatch"))
	}
	return &UserInfo{
		Username: name,
		Email:    fmt.Sprintf("%s@test.local", name),
		Subject:  "subject",
		Issuer:   "issuer",
	}, nil

}

// SetHardcodeVerifierForTest overwrite the default secret manager for testing.
// Be reminded this is for testing only.
func SetHardcodeVerifierForTest(s string) {
	m = &fakeVerifier{s}
}
func mockPopulateGroups(groupNames []string) ([]int, error) {
	res := make([]int, 0)
	for _, g := range groupNames {
		id, err := strconv.Atoi(g)
		if err != nil {
			return res, err
		}
		res = append(res, id)
	}
	return res, nil
}
