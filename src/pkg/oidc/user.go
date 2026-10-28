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
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	// maxUsernameLen is the length of the username column of the harbor_user table, in characters
	maxUsernameLen = 255
	// maxSubIssLen is the length of the subiss column of the oidc_user table, in characters
	maxSubIssLen = 255
)

// UserComment is the comment set on the users onboarded via the OIDC provider
const UserComment = "Onboarded via OIDC provider"

// ValidateUsername checks a username chosen for a new OIDC user against Harbor's username rules,
// so that all onboarding flows accept and reject the same usernames.
func ValidateUsername(username string) error {
	if utils.IsIllegalLength(username, 1, maxUsernameLen) {
		return errors.BadRequestError(nil).WithMessage("username with illegal length")
	}
	if strings.ContainsAny(username, common.IllegalCharsInUsername) {
		return errors.BadRequestError(nil).WithMessagef("username %v contains illegal characters: %v", username, common.IllegalCharsInUsername)
	}
	return nil
}

// EncryptSecretAndToken generates a new CLI secret and encrypts it together with the token for the OIDC user meta.
// The stored token is what [VerifySecret] checks, and refreshes, whenever the CLI secret is used.
func EncryptSecretAndToken(tokenBytes []byte) (string, string, error) {
	key, err := keyLoader.encryptKey()
	if err != nil {
		return "", "", err
	}
	token, err := utils.ReversibleEncrypt((string)(tokenBytes), key)
	if err != nil {
		return "", "", err
	}
	str := utils.GenerateRandomString()
	secret, err := utils.ReversibleEncrypt(str, key)
	if err != nil {
		return "", "", err
	}
	return secret, token, nil
}

// DecryptToken decrypts a token stored by [EncryptSecretAndToken], e.g. to check whether it has a refresh token
// before replacing it.
func DecryptToken(encToken string) (*Token, error) {
	key, err := keyLoader.encryptKey()
	if err != nil {
		return nil, err
	}
	tokenStr, err := utils.ReversibleDecrypt(encToken, key)
	if err != nil {
		return nil, err
	}
	token := &Token{}
	if err := json.Unmarshal([]byte(tokenStr), token); err != nil {
		return nil, err
	}
	return token, nil
}

// NewOnboardUser builds the user to pass to [github.com/goharbor/harbor/src/controller/user.Controller.OnboardOIDCUser],
// linked to the OIDC identity by subject and issuer, and with a newly generated CLI secret.
func NewOnboardUser(info *UserInfo, username string, tokenBytes []byte) (*models.User, error) {
	subIss := info.Subject + info.Issuer
	if utf8.RuneCountInString(subIss) > maxSubIssLen {
		return nil, errors.BadRequestError(nil).WithMessagef("the subject and issuer of the ID token are longer than %d characters combined: %s",
			maxSubIssLen, info.Subject)
	}
	s, t, err := EncryptSecretAndToken(tokenBytes)
	if err != nil {
		return nil, err
	}
	return &models.User{
		Username: username,
		Realname: username,
		Email:    info.Email,
		OIDCUserMeta: &models.OIDCUser{
			SubIss: subIss,
			Secret: s,
			Token:  t,
		},
		Comment: UserComment,
	}, nil
}
