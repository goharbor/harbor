// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oidc

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

// This is for testing only
type fakeVerifier struct {
	secret string
}

func (fv *fakeVerifier) VerifySecret(_ context.Context, name string, secret string) (*UserInfo, error) {
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
