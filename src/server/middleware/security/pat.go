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

package security

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils"
	pat_ctl "github.com/goharbor/harbor/src/controller/pat"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/pat/model"
)

const patPrefix = "hbr_pat_"

type pat struct{}

func (p *pat) Generate(req *http.Request) security.Context {
	ctx := req.Context()
	log := log.G(ctx)

	username, secret, ok := req.BasicAuth()
	if !ok {
		log.Debugf("PAT middleware: no basic auth found")
		return nil
	}

	log.Debugf("PAT middleware: got username=%s, secret prefix=%s", username, secret[:min(10, len(secret))])

	// Skip robot accounts - they are handled by the robot middleware
	if strings.HasPrefix(username, config.RobotPrefix(ctx)) {
		log.Debugf("PAT middleware: skipping robot account")
		return nil
	}

	// Check if this is a PAT (new tokens have the prefix, legacy tokens don't but are handled separately)
	isNewPAT := strings.HasPrefix(secret, patPrefix)
	if !isNewPAT {
		log.Debugf("PAT middleware: secret doesn't have PAT prefix, skipping")
		// For legacy PATs (migrated CLI secrets), we could handle them here
		// For now, skip and let oidcCli or other handlers deal with it
		return nil
	}

	log.Debugf("PAT middleware: verified PAT prefix, looking up user=%s", username)

	// Lookup the user
	u, err := user.Ctl.GetByName(ctx, username)
	if err != nil {
		log.Debugf("failed to get user %s for PAT verification: %v", username, err)
		return nil
	}

	log.Debugf("PAT middleware: found user ID=%d", u.UserID)

	// Remove the prefix from the secret for comparison
	secretWithoutPrefix := strings.TrimPrefix(secret, patPrefix)

	// Query all non-disabled, non-legacy PATs for this user
	pats, err := pat_ctl.Ctl.List(ctx, q.New(q.KeyWords{"user_id": u.UserID, "disabled": false, "is_legacy": false}))
	if err != nil {
		log.Debugf("failed to list PATs for user %d: %v", u.UserID, err)
		return nil
	}

	log.Debugf("PAT middleware: found %d PATs for user %d", len(pats), u.UserID)

	now := time.Now().Unix()

	// Try to find a matching PAT
	for _, token := range pats {
		// Check expiry
		if token.ExpiresAt != -1 && token.ExpiresAt <= now {
			continue
		}

		// Verify the secret
		hashedSecret := utils.Encrypt(secretWithoutPrefix, token.Salt, utils.SHA256)
		if hashedSecret != token.Secret {
			continue
		}

		// Found a matching token - update last_used_at in the background
		go func(t *model.PersonalAccessToken) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = pat_ctl.Ctl.Update(bgCtx, t, "last_used_at")
		}(token)

		log.Debugf("PAT authentication successful for user %s", username)
		return local.NewSecurityContext(u)
	}

	log.Debugf("failed to authenticate with PAT for user %s", username)
	return nil
}
