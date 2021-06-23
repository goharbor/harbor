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

package ldap

import (
	"context"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/ldap"
	"github.com/goharbor/harbor/src/pkg/ldap/model"
)

var (
	// Ctl Global instance of the LDAP controller
	Ctl = NewController()
)

// Controller define the operations related to LDAP
type Controller interface {
	// Ping test the ldap config
	Ping(ctx context.Context, cfg models.LdapConf) (bool, error)
	// SearchUser search ldap user with name
	SearchUser(ctx context.Context, username string) ([]model.User, error)
	// ImportUser import ldap users to harbor
	ImportUser(ctx context.Context, importUsers []string) ([]model.FailedImportUser, error)
	// SearchGroup search ldap group by name or by dn
	SearchGroup(ctx context.Context, groupName, groupDN string) ([]model.Group, error)
	// Create ldap session with system config
	Session(ctx context.Context) (*ldap.Session, error)
}

type controller struct {
	mgr ldap.Manager
}

// NewController ...
func NewController() Controller {
	return &controller{mgr: ldap.Mgr}
}

func (c *controller) Session(ctx context.Context) (*ldap.Session, error) {
	cfg, groupCfg, err := c.ldapConfigs(ctx)
	if err != nil {
		return nil, err
	}
	return ldap.NewSession(*cfg, *groupCfg), nil
}

func (c *controller) Ping(ctx context.Context, cfg models.LdapConf) (bool, error) {
	if len(cfg.SearchPassword) == 0 {
		pwd, err := defaultPassword(ctx)
		if err != nil {
			return false, err
		}
		if len(pwd) == 0 {
			return false, ldap.ErrEmptyPassword
		}
		cfg.SearchPassword = pwd
	}
	return c.mgr.Ping(ctx, cfg)
}

func (c *controller) ldapConfigs(ctx context.Context) (*models.LdapConf, *models.GroupConf, error) {
	cfg, err := config.LDAPConf(ctx)
	if err != nil {
		return nil, nil, err
	}
	groupCfg, err := config.LDAPGroupConf(ctx)
	if err != nil {
		log.Warningf("failed to get the ldap group config, error %v", err)
		groupCfg = &models.GroupConf{}
	}
	return cfg, groupCfg, nil
}

func (c *controller) SearchUser(ctx context.Context, username string) ([]model.User, error) {
	cfg, groupCfg, err := c.ldapConfigs(ctx)
	if err != nil {
		return nil, err
	}
	return c.mgr.SearchUser(ctx, ldap.NewSession(*cfg, *groupCfg), username)
}

func defaultPassword(ctx context.Context) (string, error) {
	mod, err := config.AuthMode(ctx)
	if err != nil {
		return "", err
	}
	if mod == common.LDAPAuth {
		conf, err := config.LDAPConf(ctx)
		if err != nil {
			return "", err
		}
		if len(conf.SearchPassword) == 0 {
			return "", ldap.ErrEmptyPassword
		}
		return conf.SearchPassword, nil
	}
	return "", ldap.ErrEmptyPassword
}

func (c *controller) ImportUser(ctx context.Context, ldapImportUsers []string) ([]model.FailedImportUser, error) {
	cfg, groupCfg, err := c.ldapConfigs(ctx)
	if err != nil {
		return nil, err
	}
	return c.mgr.ImportUser(ctx, ldap.NewSession(*cfg, *groupCfg), ldapImportUsers)
}

func (c *controller) SearchGroup(ctx context.Context, groupName, groupDN string) ([]model.Group, error) {
	cfg, groupCfg, err := c.ldapConfigs(ctx)
	if err != nil {
		return nil, err
	}
	return c.mgr.SearchGroup(ctx, ldap.NewSession(*cfg, *groupCfg), groupName, groupDN)
}
