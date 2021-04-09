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

// Package config provide config for core api and other modules
// Before accessing user settings, need to call Load()
// For system settings, no need to call Load()
package config

import (
	"context"
	"errors"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/config"
	"github.com/goharbor/harbor/src/pkg/encrypt"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/ldap/model"
)

const (
	defaultKeyPath                     = "/etc/core/key"
	defaultRegistryTokenPrivateKeyPath = "/etc/core/private_key.pem"

	// SessionCookieName is the name of the cookie for session ID
	SessionCookieName = "sid"
)

var (
	// SecretStore manages secrets
	SecretStore *secret.Store
	keyProvider encrypt.KeyProvider
	// defined as a var for testing.
	defaultCACertPath = "/etc/core/ca/ca.crt"
)

// Init configurations
func Init() {
	// init key provider
	initKeyProvider()
	log.Info("init secret store")
	// init secret store
	initSecretStore()
}

// InitWithSettings init config with predefined configs, and optionally overwrite the keyprovider
func InitWithSettings(cfgs map[string]interface{}, kp ...encrypt.KeyProvider) {
	Init()
	Ctl = NewInMemoryController()
	mgr := Ctl.GetManager()
	mgr.UpdateConfig(backgroundCtx, cfgs)
	if len(kp) > 0 {
		keyProvider = kp[0]
	}
}

// GetCfgManager return the current config manager
func GetCfgManager(ctx context.Context) config.Manager {
	return Ctl.GetManager()
}

// Load configurations
func Load(ctx context.Context) error {
	return Ctl.Load(ctx)
}

// Upload save all configurations, used by testing
func Upload(cfg map[string]interface{}) error {
	mgr := Ctl.GetManager()
	return mgr.UpdateConfig(orm.Context(), cfg)
}

// GetSystemCfg returns the system configurations
func GetSystemCfg(ctx context.Context) (map[string]interface{}, error) {
	sysCfg, err := Ctl.AllConfigs(ctx)
	if err != nil {
		return nil, err
	}
	if len(sysCfg) == 0 {
		return nil, errors.New("can not load system config, the database might be down")
	}
	return sysCfg, nil
}

// AuthMode ...
func AuthMode(ctx context.Context) (string, error) {
	err := Ctl.Load(ctx)
	if err != nil {
		log.Errorf("failed to load config, error %v", err)
		return "db_auth", err
	}
	return Ctl.GetString(ctx, common.AUTHMode), nil
}

// LDAPConf returns the setting of ldap server
func LDAPConf(ctx context.Context) (*model.LdapConf, error) {
	err := Ctl.Load(ctx)
	if err != nil {
		return nil, err
	}
	return &model.LdapConf{
		URL:               Ctl.GetString(ctx, common.LDAPURL),
		SearchDn:          Ctl.GetString(ctx, common.LDAPSearchDN),
		SearchPassword:    Ctl.GetString(ctx, common.LDAPSearchPwd),
		BaseDn:            Ctl.GetString(ctx, common.LDAPBaseDN),
		UID:               Ctl.GetString(ctx, common.LDAPUID),
		Filter:            Ctl.GetString(ctx, common.LDAPFilter),
		Scope:             Ctl.GetInt(ctx, common.LDAPScope),
		ConnectionTimeout: Ctl.GetInt(ctx, common.LDAPTimeout),
		VerifyCert:        Ctl.GetBool(ctx, common.LDAPVerifyCert),
	}, nil
}

// LDAPGroupConf returns the setting of ldap group search
func LDAPGroupConf(ctx context.Context) (*model.GroupConf, error) {
	err := Ctl.Load(ctx)
	if err != nil {
		return nil, err
	}
	return &model.GroupConf{
		BaseDN:              Ctl.GetString(ctx, common.LDAPGroupBaseDN),
		Filter:              Ctl.GetString(ctx, common.LDAPGroupSearchFilter),
		NameAttribute:       Ctl.GetString(ctx, common.LDAPGroupAttributeName),
		SearchScope:         Ctl.GetInt(ctx, common.LDAPGroupSearchScope),
		AdminDN:             Ctl.GetString(ctx, common.LDAPGroupAdminDn),
		MembershipAttribute: Ctl.GetString(ctx, common.LDAPGroupMembershipAttribute),
	}, nil
}

// TokenExpiration returns the token expiration time (in minute)
func TokenExpiration(ctx context.Context) (int, error) {
	return Ctl.GetInt(ctx, common.TokenExpiration), nil
}

// RobotTokenDuration returns the token expiration time of robot account (in minute)
func RobotTokenDuration(ctx context.Context) int {
	return Ctl.GetInt(ctx, common.RobotTokenDuration)
}

// SelfRegistration returns the enablement of self registration
func SelfRegistration(ctx context.Context) (bool, error) {
	return Ctl.GetBool(ctx, common.SelfRegistration), nil
}

// OnlyAdminCreateProject returns the flag to restrict that only sys admin can create project
func OnlyAdminCreateProject(ctx context.Context) (bool, error) {
	err := Ctl.Load(ctx)
	if err != nil {
		return true, err
	}
	return Ctl.GetString(ctx, common.ProjectCreationRestriction) == common.ProCrtRestrAdmOnly, nil
}

// Email returns email server settings
func Email(ctx context.Context) (*cfgModels.Email, error) {
	err := Ctl.Load(ctx)
	if err != nil {
		return nil, err
	}
	return &cfgModels.Email{
		Host:     Ctl.GetString(ctx, common.EmailHost),
		Port:     Ctl.GetInt(ctx, common.EmailPort),
		Username: Ctl.GetString(ctx, common.EmailUsername),
		Password: Ctl.GetString(ctx, common.EmailPassword),
		SSL:      Ctl.GetBool(ctx, common.EmailSSL),
		From:     Ctl.GetString(ctx, common.EmailFrom),
		Identity: Ctl.GetString(ctx, common.EmailIdentity),
		Insecure: Ctl.GetBool(ctx, common.EmailInsecure),
	}, nil
}

// UAASettings returns the UAASettings to access UAA service.
func UAASettings(ctx context.Context) (*models.UAASettings, error) {
	err := Ctl.Load(ctx)
	if err != nil {
		return nil, err
	}
	us := &models.UAASettings{
		Endpoint:     Ctl.GetString(ctx, common.UAAEndpoint),
		ClientID:     Ctl.GetString(ctx, common.UAAClientID),
		ClientSecret: Ctl.GetString(ctx, common.UAAClientSecret),
		VerifyCert:   Ctl.GetBool(ctx, common.UAAVerifyCert),
	}
	return us, nil
}

// ReadOnly returns a bool to indicates if Harbor is in read only mode.
func ReadOnly(ctx context.Context) bool {
	return Ctl.GetBool(ctx, common.ReadOnly)
}

// HTTPAuthProxySetting returns the setting of HTTP Auth proxy.  the settings are only meaningful when the auth_mode is
// set to http_auth
func HTTPAuthProxySetting(ctx context.Context) (*cfgModels.HTTPAuthProxy, error) {
	if err := Ctl.Load(ctx); err != nil {
		return nil, err
	}
	return &cfgModels.HTTPAuthProxy{
		Endpoint:            Ctl.GetString(ctx, common.HTTPAuthProxyEndpoint),
		TokenReviewEndpoint: Ctl.GetString(ctx, common.HTTPAuthProxyTokenReviewEndpoint),
		AdminGroups:         splitAndTrim(Ctl.GetString(ctx, common.HTTPAuthProxyAdminGroups), ","),
		AdminUsernames:      splitAndTrim(Ctl.GetString(ctx, common.HTTPAuthProxyAdminUsernames), ","),
		VerifyCert:          Ctl.GetBool(ctx, common.HTTPAuthProxyVerifyCert),
		SkipSearch:          Ctl.GetBool(ctx, common.HTTPAuthProxySkipSearch),
		ServerCertificate:   Ctl.GetString(ctx, common.HTTPAuthProxyServerCertificate),
	}, nil
}

// OIDCSetting returns the setting of OIDC provider, currently there's only one OIDC provider allowed for Harbor and it's
// only effective when auth_mode is set to oidc_auth
func OIDCSetting(ctx context.Context) (*cfgModels.OIDCSetting, error) {
	if err := Ctl.Load(ctx); err != nil {
		return nil, err
	}
	scopeStr := Ctl.GetString(ctx, common.OIDCScope)
	extEndpoint := strings.TrimSuffix(Ctl.GetString(nil, common.ExtEndpoint), "/")
	scope := splitAndTrim(scopeStr, ",")
	return &cfgModels.OIDCSetting{
		Name:               Ctl.GetString(ctx, common.OIDCName),
		Endpoint:           Ctl.GetString(ctx, common.OIDCEndpoint),
		VerifyCert:         Ctl.GetBool(ctx, common.OIDCVerifyCert),
		AutoOnboard:        Ctl.GetBool(ctx, common.OIDCAutoOnboard),
		ClientID:           Ctl.GetString(ctx, common.OIDCCLientID),
		ClientSecret:       Ctl.GetString(ctx, common.OIDCClientSecret),
		GroupsClaim:        Ctl.GetString(ctx, common.OIDCGroupsClaim),
		AdminGroup:         Ctl.GetString(ctx, common.OIDCAdminGroup),
		RedirectURL:        extEndpoint + common.OIDCCallbackPath,
		Scope:              scope,
		UserClaim:          Ctl.GetString(ctx, common.OIDCUserClaim),
		ExtraRedirectParms: Ctl.Get(ctx, common.OIDCExtraRedirectParms).GetStringToStringMap(),
	}, nil
}

// NotificationEnable returns a bool to indicates if notification enabled in harbor
func NotificationEnable(ctx context.Context) bool {
	return Ctl.GetBool(ctx, common.NotificationEnable)
}

// QuotaPerProjectEnable returns a bool to indicates if quota per project enabled in harbor
func QuotaPerProjectEnable(ctx context.Context) bool {
	return Ctl.GetBool(ctx, common.QuotaPerProjectEnable)
}

// QuotaSetting returns the setting of quota.
func QuotaSetting(ctx context.Context) (*cfgModels.QuotaSetting, error) {
	if err := Ctl.Load(ctx); err != nil {
		return nil, err
	}
	return &cfgModels.QuotaSetting{
		StoragePerProject: Ctl.Get(ctx, common.StoragePerProject).GetInt64(),
	}, nil
}

// RobotPrefix user defined robot name prefix.
func RobotPrefix(ctx context.Context) string {
	return Ctl.GetString(ctx, common.RobotNamePrefix)
}

func splitAndTrim(s, sep string) []string {
	res := make([]string, 0)
	for _, s := range strings.Split(s, sep) {
		if e := strings.TrimSpace(s); len(e) > 0 {
			res = append(res, e)
		}
	}
	return res
}
