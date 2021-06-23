//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package validate

import (
	"context"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/errors"
)

// LdapGroupValidateRule validate the ldap group configuration
type LdapGroupValidateRule struct {
}

// Validate validate the ldap group config settings, cfgs is the config items need to be updated, the final config should be merged with the cfgMgr
func (l LdapGroupValidateRule) Validate(ctx context.Context, cfgMgr config.Manager, cfgs map[string]interface{}) error {
	cfg := &cfgModels.GroupConf{
		Filter:              cfgMgr.Get(ctx, common.LDAPGroupSearchFilter).GetString(),
		NameAttribute:       cfgMgr.Get(ctx, common.LDAPGroupAttributeName).GetString(),
		MembershipAttribute: cfgMgr.Get(ctx, common.LDAPGroupMembershipAttribute).GetString(),
	}
	updated := false
	// Merge the cfgs and the cfgMgr to get the final GroupConf
	if val, exist := cfgs[common.LDAPGroupSearchFilter]; exist {
		cfg.Filter = val.(string)
		updated = true
	}
	if val, exist := cfgs[common.LDAPGroupAttributeName]; exist {
		cfg.NameAttribute = val.(string)
		updated = true
	}
	if val, exist := cfgs[common.LDAPGroupMembershipAttribute]; exist {
		cfg.MembershipAttribute = val.(string)
		updated = true
	}
	if !updated {
		return nil
	}

	if len(cfg.Filter) == 0 {
		// skip to validate group config
		return nil
	}
	if len(cfg.NameAttribute) == 0 {
		return errors.New("ldap group name attribute can not be empty")
	}
	if len(cfg.MembershipAttribute) == 0 {
		return errors.New("ldap group membership attribute can not be empty")
	}
	return nil
}
