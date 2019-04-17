package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/pkg/errors"
	"sync"
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
	// SetSecret sets the secret and token based on the ID of the user, when setting the secret the user has to be
	// onboarded to Harbor DB.
	SetSecret(userID int, secret string, token *Token) error
	// VerifySecret verifies the secret and the token associated with it, it refreshes the token in the DB if it's
	// refreshed during the verification
	VerifySecret(ctx context.Context, userID int, secret string) error
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

// SetSecret sets the secret and token based on the ID of the user, when setting the secret the user has to be
// onboarded to Harbor DB.
func (dm *defaultManager) SetSecret(userID int, secret string, token *Token) error {
	key, err := dm.getEncryptKey()
	if err != nil {
		return fmt.Errorf("failed to load the key for encryption/decryption： %v", err)
	}
	oidcUser, err := dao.GetOIDCUserByUserID(userID)
	if oidcUser == nil {
		return fmt.Errorf("failed to get oidc user info, error: %v", err)
	}
	encSecret, _ := utils.ReversibleEncrypt(secret, key)
	tb, _ := json.Marshal(token)
	encToken, _ := utils.ReversibleEncrypt(string(tb), key)
	oidcUser.Secret = encSecret
	oidcUser.Token = encToken
	return dao.UpdateOIDCUser(oidcUser)
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
	tokenStr, err := utils.ReversibleDecrypt(oidcUser.Token, key)
	if err != nil {
		return verifyError(err)
	}
	token := &Token{}
	err = json.Unmarshal(([]byte)(tokenStr), token)
	if err != nil {
		return verifyError(err)
	}
	_, err = VerifyToken(ctx, token.IDToken)
	if err == nil {
		return nil
	}
	log.Infof("Failed to verify ID Token, error: %v, refreshing...", err)
	t, err := RefreshToken(ctx, token)
	if err != nil {
		return verifyError(err)
	}
	err = dm.SetSecret(oidcUser.UserID, secret, t)
	if err != nil {
		log.Warningf("Failed to update the token in DB: %v, ignore this error.", err)
	}
	return nil
}

// VerifySecret verifies the secret and the token associated with it, it tries to update the token in the DB if it's
// refreshed during the verification
func VerifySecret(ctx context.Context, userID int, secret string) error {
	return m.VerifySecret(ctx, userID, secret)
}
