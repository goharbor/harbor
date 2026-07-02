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

package migration

import (
	"context"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/pat"
	pat_model "github.com/goharbor/harbor/src/pkg/pat/model"
)

// oidcUserRow holds OIDC user data for migration
type oidcUserRow struct {
	ID     int64  `orm:"column(id)"`
	UserID int64  `orm:"column(user_id)"`
	Secret string `orm:"column(secret)"`
}

// MigrateCliSecretsToLegacyPATs converts existing OIDC CLI secrets to legacy PAT records.
// Safe to run multiple times — skips users already having a PAT named "cli-secret".
func MigrateCliSecretsToLegacyPATs(ctx context.Context) error {
	logger := log.G(ctx)
	logger.Infof("starting migration of OIDC CLI secrets to legacy PATs")

	// Get the ORM from context
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		logger.Errorf("failed to get ORM from context: %v", err)
		return err
	}

	// Fetch all OIDC users with non-empty secrets
	var oidcUsers []oidcUserRow
	_, err = ormer.Raw("SELECT id, user_id, secret FROM oidc_user WHERE secret IS NOT NULL AND secret != ''").QueryRows(&oidcUsers)
	if err != nil {
		logger.Errorf("failed to query oidc_user table: %v", err)
		return err
	}

	if len(oidcUsers) == 0 {
		logger.Infof("no OIDC CLI secrets found to migrate")
		return nil
	}

	logger.Infof("found %d OIDC CLI secrets to migrate", len(oidcUsers))

	patMgr := pat.NewManager()
	migratedCount := 0
	skippedCount := 0
	errorCount := 0

	secretKey, err := config.SecretKey()
	if err != nil {
		logger.Errorf("failed to get secret key: %v", err)
		return err
	}

	for _, row := range oidcUsers {
		userID := int(row.UserID)
		encryptedSecret := row.Secret

		// Check if a PAT named "cli-secret" already exists for this user
		existing, err := patMgr.List(ctx, q.New(q.KeyWords{"user_id": userID, "name": "cli-secret"}))
		if err != nil {
			logger.Warningf("failed to check existing PAT for user %d: %v", userID, err)
			errorCount++
			continue
		}
		if len(existing) > 0 {
			logger.Debugf("user %d already has a cli-secret PAT, skipping", userID)
			skippedCount++
			continue
		}

		// Decrypt the AES CLI secret
		plaintext, err := utils.ReversibleDecrypt(encryptedSecret, secretKey)
		if err != nil {
			logger.Warningf("could not decrypt CLI secret for user %d, skipping migration: %v", userID, err)
			errorCount++
			continue
		}

		// Re-hash the plaintext as PBKDF2 (same as robot accounts)
		salt := utils.GenerateRandomString()
		hashed := utils.Encrypt(plaintext, salt, utils.SHA256)

		// Create a legacy PAT record
		legacyPAT := &pat_model.PersonalAccessToken{
			UserID:      userID,
			Name:        "cli-secret",
			Description: "Migrated from OIDC CLI secret",
			Secret:      hashed,
			Salt:        salt,
			ExpiresAt:   -1,
			Disabled:    false,
			IsLegacy:    true,
		}

		_, err = patMgr.Create(ctx, legacyPAT)
		if err != nil {
			logger.Warningf("failed to create legacy PAT for user %d: %v", userID, err)
			errorCount++
			continue
		}

		migratedCount++
		logger.Debugf("migrated CLI secret for user %d to legacy PAT", userID)
	}

	logger.Infof("CLI secrets migration complete: migrated=%d, skipped=%d, errors=%d",
		migratedCount, skippedCount, errorCount)

	if errorCount > 0 {
		return errors.Errorf("migration completed with %d errors", errorCount)
	}

	return nil
}
