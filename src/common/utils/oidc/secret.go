package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
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
	// VerifySecret verifies the secret and the token associated with it, it refreshes the token in the DB if it's
	// refreshed during the verification.  It returns a populated user model based on the ID token associated with the secret.
	VerifySecret(ctx context.Context, username string, secret string) (*models.User, error)
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

// VerifySecret verifies the secret and the token associated with it, it refreshes the token in the DB if it's
// refreshed during the verification.  It returns a populated user model based on the ID token associated with the secret.
func (dm *defaultManager) VerifySecret(ctx context.Context, username string, secret string) (*models.User, error) {
	user, err := dao.GetUser(models.User{Username: username})
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, verifyError(fmt.Errorf("user does not exist, name: %s", username))
	}
	oidcUser, err := dao.GetOIDCUserByUserID(user.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oidc user info, error: %v", err)
	}
	if oidcUser == nil {
		return nil, fmt.Errorf("user is not onboarded as OIDC user")
	}
	key, err := dm.getEncryptKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load the key for encryption/decryption： %v", err)
	}
	plainSecret, err := utils.ReversibleDecrypt(oidcUser.Secret, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret from DB: %v", err)
	}
	if secret != plainSecret {
		return nil, verifyError(errors.New("secret mismatch"))
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
		token, err = refreshToken(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token")
		}
		tb, err := json.Marshal(token)
		if err != nil {
			return nil, fmt.Errorf("failed to encode the refreshed token, error: %v", err)
		}
		encToken, _ := utils.ReversibleEncrypt(string(tb), key)
		oidcUser.Token = encToken
		err = dao.UpdateOIDCUser(oidcUser)
		if err != nil {
			log.Errorf("Failed to persist token, user id: %d, error: %v", oidcUser.UserID, err)
		}
		log.Debug("Token refreshed and persisted")
	}
	info, err := UserInfoFromToken(ctx, token)
	if err != nil {
		return nil, verifyError(err)
	}
	gids, err := group.PopulateGroup(models.UserGroupsFromName(info.Groups, common.OIDCGroupType))
	if err != nil {
		log.Warningf("failed to get group ID, error: %v, skip populating groups", err)
	} else {
		user.GroupIDs = gids
	}
	return user, nil
}

// VerifySecret calls the manager to verify the secret.
func VerifySecret(ctx context.Context, name string, secret string) (*models.User, error) {
	return m.VerifySecret(ctx, name, secret)
}
