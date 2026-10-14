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
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/oidc/dao"
)

// SecretVerifyError wraps the different errors happened when verifying a secret for OIDC user.  When seeing this error,
// the caller should consider this an authentication error.
type SecretVerifyError struct {
	cause error
}

func (se *SecretVerifyError) Error() string {
	return fmt.Sprintf("failed to verify the secret: %v", se.cause)
}

func verifyError(err error) error {
	return &SecretVerifyError{err}
}

// SecretManager is the interface for store and verify the secret
type SecretManager interface {
	// VerifySecret verifies the secret and the token associated with it, it refreshes the token in the DB if it's
	// refreshed during the verification.
	VerifySecret(ctx context.Context, username string, secret string) (*UserInfo, error)
}

type keyGetter struct {
	sync.RWMutex
	key string
}

func (kg *keyGetter) encryptKey() (string, error) {
	kg.RLock()
	if kg.key == "" {
		kg.RUnlock()
		kg.Lock()
		defer kg.Unlock()
		if kg.key == "" {
			k, err := config.SecretKey()
			if err != nil {
				return "", err
			}
			kg.key = k
		}
	} else {
		defer kg.RUnlock()
	}
	return kg.key, nil
}

var keyLoader = &keyGetter{}

type defaultManager struct {
	metaDao dao.MetaDAO
}

var m SecretManager = &defaultManager{
	metaDao: dao.NewMetaDao(),
}

// VerifySecret verifies the secret and the token associated with it, it refreshes the token in the DB if it's
// refreshed during the verification.  It returns a populated user model based on the ID token associated with the secret.
func (dm *defaultManager) VerifySecret(ctx context.Context, username string, secret string) (*UserInfo, error) {
	log.Debugf("Verifying the secret for user: %s", username)
	oidcUser, err := dm.metaDao.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get oidc user info, error: %v", err)
	}
	if oidcUser == nil {
		return nil, fmt.Errorf("user is not onboarded as OIDC user, username: %s", username)
	}
	key, err := keyLoader.encryptKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load the key for encryption/decryption： %v", err)
	}
	plainSecret, err := utils.ReversibleDecrypt(oidcUser.Secret, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret from DB: %v", err)
	}
	if subtle.ConstantTimeCompare([]byte(secret), []byte(plainSecret)) != 1 {
		return nil, verifyError(fmt.Errorf("secret mismatch, username: %s", username))
	}
	tokenStr, err := utils.ReversibleDecrypt(oidcUser.Token, key)
	if err != nil {
		return nil, verifyError(err)
	}
	token := &Token{}
	err = json.Unmarshal(([]byte)(tokenStr), token)
	if err != nil {
		return nil, verifyError(err)
	}
	if !token.Valid() {
		log.Debug("Refreshing token")
		refreshed, err := refreshTokenFn(ctx, token)
		if err != nil {
			// The CLI secret is a user-owned credential, its lifetime is not bound to the IdP session.
			// Keep the stored token and use the last known user info.
			warnRefreshFailure(username, err)
			if token.UserInfo != nil {
				return token.UserInfo, nil
			}
			info, err := userInfoFromStoredTokenFn(ctx, token)
			if err != nil {
				return nil, verifyError(err)
			}
			return info, nil
		}
		refreshFailures.Delete(username)
		// a refresh response may omit the ID token, keep the previous one
		if refreshed.RawIDToken == "" {
			refreshed.RawIDToken = token.RawIDToken
		}
		refreshed.UserInfo = token.UserInfo
		token = refreshed
		dm.persistToken(ctx, oidcUser, token, key)
		log.Debug("Token refreshed and persisted")
	}
	info, err := userInfoFromTokenFn(ctx, token)
	if err != nil {
		return nil, verifyError(err)
	}
	if !sameAuthorization(token.UserInfo, info) {
		token.UserInfo = info
		dm.persistToken(ctx, oidcUser, token, key)
	}
	return info, nil
}

// persistToken only updates the token column of the record
func (dm *defaultManager) persistToken(ctx context.Context, oidcUser *models.OIDCUser, token *Token, key string) {
	tb, err := json.Marshal(token)
	if err != nil {
		log.Errorf("Failed to encode token, user id: %d, error: %v", oidcUser.UserID, err)
		return
	}
	encToken, err := utils.ReversibleEncrypt(string(tb), key)
	if err != nil {
		log.Errorf("Failed to encrypt token, user id: %d, error: %v", oidcUser.UserID, err)
		return
	}
	oidcUser.Token = encToken
	if err := dm.metaDao.Update(ctx, oidcUser, "token"); err != nil {
		log.Errorf("Failed to persist token, user id: %d, error: %v", oidcUser.UserID, err)
	}
}

// sameAuthorization reports whether the stored user info grants the same groups and admin status
func sameAuthorization(stored, current *UserInfo) bool {
	return stored != nil && current != nil && stored.AdminGroupMember == current.AdminGroupMember &&
		slices.Equal(stored.Groups, current.Groups)
}

var (
	refreshTokenFn            = refreshToken
	userInfoFromTokenFn       = UserInfoFromToken
	userInfoFromStoredTokenFn = userInfoFromStoredToken
	// refreshFailures records the users whose token refresh failed, so the failure is logged once per user
	// instead of once per request.  An entry is removed when a refresh for that user succeeds.
	refreshFailures sync.Map
)

func warnRefreshFailure(username string, err error) {
	if _, logged := refreshFailures.LoadOrStore(username, struct{}{}); logged {
		log.Debugf("Failed to refresh the OIDC token, keep using the stored token, username: %s, error: %v", username, err)
		return
	}
	log.Warningf("Failed to refresh the OIDC token, keep using the stored token until a refresh succeeds, username: %s, error: %v", username, err)
}

// userInfoFromStoredToken resolves the user from the ID token stored with the secret, without calling the
// userinfo endpoint of the OIDC provider, because the access token is expired and could not be refreshed.
// It is only used when no user info was stored with the token.
func userInfoFromStoredToken(ctx context.Context, token *Token) (*UserInfo, error) {
	if token.RawIDToken == "" {
		return nil, fmt.Errorf("the stored token is expired and has no ID token")
	}
	// preload the provider, in case it was not previously loaded by the UI
	if _, err := provider.get(ctx); err != nil {
		return nil, err
	}
	setting, err := config.OIDCSetting(ctx)
	if err != nil {
		return nil, err
	}
	return UserInfoFromIDToken(ctx, token, *setting)
}

// VerifySecret calls the manager to verify the secret.
func VerifySecret(ctx context.Context, name string, secret string) (*UserInfo, error) {
	return m.VerifySecret(ctx, name, secret)
}
