package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/pkg/errors"
	"sync"
	"time"
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
	// refreshed during the verification
	VerifySecret(ctx context.Context, userID int, secret string) error
	// VerifyToken verifies the token in the model from parm,
	// and refreshes the token in the DB if it's refreshed during the verification.
	VerifyToken(ctx context.Context, user *models.OIDCUser) error
}

type defaultManager struct {
	sync.Mutex
	key string
}

var m SecretManager = &defaultManager{}

func (dm *defaultManager) getEncryptKey() (string, error) {
	if dm.key == "" {
		dm.Lock()
		defer dm.Unlock()
		if dm.key == "" {
			key, err := config.SecretKey()
			if err != nil {
				return "", err
			}
			dm.key = key
		}
	}
	return dm.key, nil
}

// VerifySecret verifies the secret and the token associated with it, it tries to update the token in the DB if it's
// refreshed during the verification
func (dm *defaultManager) VerifySecret(ctx context.Context, userID int, secret string) error {
	oidcUser, err := dao.GetOIDCUserByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get oidc user info, error: %v", err)
	}
	if oidcUser == nil {
		return fmt.Errorf("user is not onboarded as OIDC user")
	}
	key, err := dm.getEncryptKey()
	if err != nil {
		return fmt.Errorf("failed to load the key for encryption/decryption： %v", err)
	}
	plainSecret, err := utils.ReversibleDecrypt(oidcUser.Secret, key)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret from DB: %v", err)
	}
	if secret != plainSecret {
		return verifyError(errors.New("secret mismatch"))
	}
	return dm.VerifyToken(ctx, oidcUser)
}

// VerifyToken checks the expiration of the token in the model, (we'll only do expiration checks b/c according to spec,
// the response may not have ID token:
// https://openid.net/specs/openid-connect-core-1_0.html#RefreshTokenResponse
// and it will try to refresh the token
// if it's expired, if the refresh is successful it will persist the token and consider the verification successful.
func (dm *defaultManager) VerifyToken(ctx context.Context, user *models.OIDCUser) error {
	if user == nil {
		return verifyError(fmt.Errorf("input user is nil"))
	}
	key, err := dm.getEncryptKey()
	if err != nil {
		return fmt.Errorf("failed to load the key for encryption/decryption： %v", err)
	}
	tokenStr, err := utils.ReversibleDecrypt(user.Token, key)
	if err != nil {
		return verifyError(err)
	}
	token := &Token{}
	err = json.Unmarshal(([]byte)(tokenStr), token)
	if err != nil {
		return verifyError(err)
	}
	log.Debugf("Token string for verify: %s", tokenStr)
	if !token.Expiry.After(time.Now()) {
		log.Info("Token string has expired, refreshing...")
	}
	t, err := RefreshToken(ctx, token)
	if err != nil {
		return verifyError(err)
	}
	tb, err := json.Marshal(t)
	if err != nil {
		log.Warningf("Failed to encode the refreshed token, error: %v", err)
	}
	encToken, _ := utils.ReversibleEncrypt(string(tb), key)
	user.Token = encToken
	err = dao.UpdateOIDCUser(user)
	if err != nil {
		log.Warningf("Failed to update the token in DB: %v, ignore this error.", err)
	}
	return nil
}

// VerifySecret calls the manager to verify the secret.
func VerifySecret(ctx context.Context, userID int, secret string) error {
	return m.VerifySecret(ctx, userID, secret)
}

// VerifyAndPersistToken calls the manager to verify token and persist it if it's refreshed.
func VerifyAndPersistToken(ctx context.Context, user *models.OIDCUser) error {
	return m.VerifyToken(ctx, user)
}
