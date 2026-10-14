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

package config

import (
	"context"
	"slices"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	offlineAccessScope = "offline_access"
	// Harbor requests offline access from Google with access_type=offline instead of the offline_access scope
	googleEndpoint = "https://accounts.google.com"
)

// warnOIDCScope logs a warning when the updated OIDC scope does not request offline access.
func warnOIDCScope(ctx context.Context, mgr config.Manager, cfgs map[string]any) {
	scope, ok := cfgs[common.OIDCScope].(string)
	if !ok {
		return
	}
	endpoint, ok := cfgs[common.OIDCEndpoint].(string)
	if !ok {
		endpoint = mgr.Get(ctx, common.OIDCEndpoint).GetString()
	}
	if requiresOfflineAccessScope(endpoint) && !hasOfflineAccessScope(scope) {
		log.Warningf("%s %q does not include %s, the token stored with the OIDC CLI secret cannot be refreshed "+
			"once the session in the OIDC provider ends, the CLI secret keeps the last known groups and admin status", common.OIDCScope, scope, offlineAccessScope)
	}
}

func requiresOfflineAccessScope(endpoint string) bool {
	return !strings.HasPrefix(endpoint, googleEndpoint)
}

func hasOfflineAccessScope(scope string) bool {
	return slices.Contains(config.SplitAndTrim(scope, ","), offlineAccessScope)
}
