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
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/usergroup"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"

	"go.pinniped.dev/pkg/oidcclient/pkce"
)

const (
	googleEndpoint = "https://accounts.google.com"
)

type claimsProvider interface {
	Claims(v interface{}) error
}

type providerHelper struct {
	sync.Mutex
	instance     atomic.Value
	creationTime time.Time
}

func (p *providerHelper) get(ctx context.Context) (*gooidc.Provider, error) {
	if p.instance.Load() != nil {
		if time.Since(p.creationTime) > 3*time.Second {
			if err := p.create(ctx); err != nil {
				return nil, err
			}
		}
	} else {
		p.Lock()
		defer p.Unlock()
		if p.instance.Load() == nil {
			if err := p.create(ctx); err != nil {
				return nil, err
			}
		}
	}

	return p.instance.Load().(*gooidc.Provider), nil
}

func (p *providerHelper) create(ctx context.Context) error {
	s, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("Failed to get OIDC configuration, error: %v", err)
		return err
	}
	c := clientCtx(ctx, s.VerifyCert)
	provider, err := gooidc.NewProvider(c, s.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider, error: %v", err)
	}
	p.instance.Store(provider)
	p.creationTime = time.Now()
	return nil
}

var provider = &providerHelper{}

var insecureTransport = &http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
	Proxy: http.ProxyFromEnvironment,
}

// Token wraps the attributes of a oauth2 token plus the attribute of ID token
type Token struct {
	oauth2.Token
	RawIDToken string `json:"id_token,omitempty"`
}

// UserInfo wraps the information that is extracted via token.  It will be transformed to data object that is persisted
// in the DB
type UserInfo struct {
	Issuer              string   `json:"iss"`
	Subject             string   `json:"sub"`
	Username            string   `json:"name"`
	Email               string   `json:"email"`
	Groups              []string `json:"groups"`
	AdminGroupMember    bool     `json:"admin_group_member"`
	autoOnboardUsername string
	hasGroupClaim       bool
}

func getOauthConf(ctx context.Context) (*oauth2.Config, error) {
	p, err := provider.get(ctx)
	if err != nil {
		return nil, err
	}
	setting, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("Failed to get OIDC configuration, error: %v", err)
		return nil, err
	}
	scopes := make([]string, 0)
	for _, sc := range setting.Scope {
		if strings.HasPrefix(p.Endpoint().AuthURL, googleEndpoint) && sc == gooidc.ScopeOfflineAccess {
			log.Warningf("Dropped unsupported scope: %s ", sc)
			continue
		}
		scopes = append(scopes, sc)
	}
	return &oauth2.Config{
		ClientID:     setting.ClientID,
		ClientSecret: setting.ClientSecret,
		Scopes:       scopes,
		RedirectURL:  setting.RedirectURL,
		Endpoint:     p.Endpoint(),
	}, nil
}

// AuthCodeURL returns the URL for OIDC provider's consent page.
// The state should be verified when user is redirected back to Harbor.
// pkceCode is for the PKCE workflow of authentication.  It is optional.
func AuthCodeURL(ctx context.Context, state string, pkceCode pkce.Code) (string, error) {
	conf, err := getOauthConf(ctx)
	if err != nil {
		log.Errorf("Failed to get OAuth configuration, error: %v", err)
		return "", err
	}
	var options []oauth2.AuthCodeOption
	setting, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("Failed to get OIDC configuration, error: %v", err)
		return "", err
	}
	for k, v := range setting.ExtraRedirectParms {
		options = append(options, oauth2.SetAuthURLParam(k, v))
	}
	if strings.HasPrefix(conf.Endpoint.AuthURL, googleEndpoint) { // make sure the refresh token will be returned
		options = append(options, oauth2.AccessTypeOffline)
		options = append(options, oauth2.SetAuthURLParam("prompt", "consent"))
	}
	if len(pkceCode) > 0 {
		options = append(options, pkceCode.Challenge())
		options = append(options, pkceCode.Method())
	}
	return conf.AuthCodeURL(state, options...), nil
}

// ExchangeToken get the token from token provider via the code
func ExchangeToken(ctx context.Context, code string, pkceCode pkce.Code) (*Token, error) {
	oauth, err := getOauthConf(ctx)
	if err != nil {
		log.Errorf("Failed to get OAuth configuration, error: %v", err)
		return nil, err
	}
	setting, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("Failed to get OIDC configuration, error: %v", err)
		return nil, err
	}
	ctx = clientCtx(ctx, setting.VerifyCert)
	var opts []oauth2.AuthCodeOption
	if len(pkceCode) > 0 {
		opts = append(opts, pkceCode.Verifier())
	}
	oauthToken, err := oauth.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, err
	}
	return &Token{Token: *oauthToken, RawIDToken: oauthToken.Extra("id_token").(string)}, nil
}

func parseIDToken(ctx context.Context, rawIDToken string) (*gooidc.IDToken, error) {
	conf := &gooidc.Config{SkipClientIDCheck: true, SkipExpiryCheck: true}
	return verifyTokenWithConfig(ctx, rawIDToken, conf)
}

// VerifyToken verifies the ID token based on the OIDC settings
func VerifyToken(ctx context.Context, rawIDToken string) (*gooidc.IDToken, error) {
	return verifyTokenWithConfig(ctx, rawIDToken, nil)
}

func verifyTokenWithConfig(ctx context.Context, rawIDToken string, conf *gooidc.Config) (*gooidc.IDToken, error) {
	log.Debugf("Raw ID token for verification: %s", rawIDToken)
	p, err := provider.get(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("Failed to get OIDC configuration, error: %v", err)
		return nil, err
	}
	if conf == nil {
		conf = &gooidc.Config{ClientID: settings.ClientID}
	}
	verifier := p.Verifier(conf)
	ctx = clientCtx(ctx, settings.VerifyCert)
	return verifier.Verify(ctx, rawIDToken)
}

func clientCtx(ctx context.Context, verifyCert bool) context.Context {
	var client *http.Client
	if !verifyCert {
		client = &http.Client{
			Transport: insecureTransport,
		}
	} else {
		client = &http.Client{}
	}
	return gooidc.ClientContext(ctx, client)
}

// refreshToken tries to refresh the token if it's expired, if it doesn't the
// original one will be returned.
func refreshToken(ctx context.Context, token *Token) (*Token, error) {
	oauthCfg, err := getOauthConf(ctx)
	if err != nil {
		return nil, err
	}
	setting, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("Failed to get OIDC configuration, error: %v", err)
		return nil, err
	}
	cctx := clientCtx(ctx, setting.VerifyCert)
	ts := oauthCfg.TokenSource(cctx, &token.Token)
	nt, err := ts.Token()
	if err != nil {
		return nil, err
	}
	it, ok := nt.Extra("id_token").(string)
	if !ok {
		log.Debug("id_token not exist in refresh response")
	}
	return &Token{Token: *nt, RawIDToken: it}, nil
}

// UserInfoFromToken tries to call the UserInfo endpoint of the OIDC provider, and consolidate with ID token
// to generate a UserInfo object, if the ID token is not in the input token struct, some attributes will be empty
func UserInfoFromToken(ctx context.Context, token *Token) (*UserInfo, error) {
	// #10913: preload the configuration, in case it was not previously loaded by the UI
	_, err := provider.get(ctx)
	if err != nil {
		return nil, err
	}
	setting, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("Failed to get OIDC configuration, error: %v", err)
		return nil, err
	}
	local, err := UserInfoFromIDToken(ctx, token, *setting)
	if err != nil {
		return nil, err
	}
	remote, err := userInfoFromRemote(ctx, token, *setting)
	if err != nil {
		log.Warningf("Failed to get userInfo by calling remote userinfo endpoint, error: %v ", err)
	}
	if remote != nil && local != nil {
		if remote.Subject != local.Subject {
			return nil, fmt.Errorf("the subject from userinfo: %s does not match the subject from ID token: %s, probably a security attack happened", remote.Subject, local.Subject)
		}
		return mergeUserInfo(remote, local), nil
	} else if remote != nil && local == nil {
		return remote, nil
	} else if local != nil && remote == nil {
		log.Debugf("Fall back to user data from ID token.")
		return local, nil
	}
	return nil, fmt.Errorf("failed to get userinfo from both remote and ID token")
}

func mergeUserInfo(remote, local *UserInfo) *UserInfo {
	res := &UserInfo{
		// data only contained in ID token
		Subject: local.Subject,
		Issuer:  local.Issuer,
		// Used data from userinfo
		Email: remote.Email,
	}
	// priority for username (high to low):
	// 1. Username based on the auto onboard claim from ID token
	// 2. Username from response of userinfo endpoint
	// 3. Username from the default "name" claim from ID token
	if local.autoOnboardUsername != "" {
		res.Username = local.autoOnboardUsername
	} else if remote.Username != "" {
		res.Username = remote.Username
	} else {
		res.Username = local.Username
	}
	if remote.hasGroupClaim {
		res.Groups = remote.Groups
		res.AdminGroupMember = remote.AdminGroupMember
		res.hasGroupClaim = true
	} else if local.hasGroupClaim {
		res.Groups = local.Groups
		res.AdminGroupMember = local.AdminGroupMember
		res.hasGroupClaim = true
	} else {
		res.Groups = []string{}
	}
	return res
}

func userInfoFromRemote(ctx context.Context, token *Token, setting cfgModels.OIDCSetting) (*UserInfo, error) {
	p, err := provider.get(ctx)
	if err != nil {
		return nil, err
	}
	cctx := clientCtx(ctx, setting.VerifyCert)
	u, err := p.UserInfo(cctx, oauth2.StaticTokenSource(&token.Token))
	if err != nil {
		return nil, err
	}
	return userInfoFromClaims(u, setting)
}

// UserInfoFromIDToken extract user info from ID token
func UserInfoFromIDToken(ctx context.Context, token *Token, setting cfgModels.OIDCSetting) (*UserInfo, error) {
	if token.RawIDToken == "" {
		return nil, nil
	}
	idt, err := parseIDToken(ctx, token.RawIDToken)
	if err != nil {
		return nil, err
	}

	return userInfoFromClaims(idt, setting)
}

func userInfoFromClaims(c claimsProvider, setting cfgModels.OIDCSetting) (*UserInfo, error) {
	res := &UserInfo{}
	if err := c.Claims(res); err != nil {
		return nil, err
	}
	if setting.UserClaim != "" {
		allClaims := make(map[string]interface{})
		if err := c.Claims(&allClaims); err != nil {
			return nil, err
		}

		if username, ok := allClaims[setting.UserClaim].(string); ok {
			// res.Username and autoOnboardUsername both need to be set to create a fallback when mergeUserInfo has not been successfully called.
			// This can for example occur when remote fails and only a local token is available for onboarding.
			// Otherwise the onboard flow only has a fallback when "name" is set in the token, which is not always the case as a custom Username Claim could be configured.
			res.autoOnboardUsername, res.Username = username, username
		} else {
			log.Warningf("OIDC. Failed to recover Username from claim. Claim '%s' is invalid or not a string", setting.UserClaim)
		}
	}
	res.Groups, res.hasGroupClaim = groupsFromClaims(c, setting.GroupsClaim)
	if len(setting.AdminGroup) > 0 {
		for _, g := range res.Groups {
			if g == setting.AdminGroup {
				res.AdminGroupMember = true
				break
			}
		}
	}
	return res, nil
}

// groupsFromClaims fetches the group name list from claimprovider, such as decoded ID token.
// If the claims does not have the claim defined as k, the second return value will be false, otherwise true
func groupsFromClaims(gp claimsProvider, k string) ([]string, bool) {
	res := make([]string, 0)
	claimMap := make(map[string]interface{})
	if err := gp.Claims(&claimMap); err != nil {
		log.Errorf("failed to fetch claims, error: %v", err)
		return res, false
	}
	g, ok := claimMap[k].([]interface{})
	if !ok {
		if len(strings.TrimSpace(k)) > 0 {
			log.Warningf("Unable to get groups from claims, claims: %+v, groups claims key: %s", claimMap, k)
		}
		return res, false
	}
	for _, e := range g {
		s, ok := e.(string)
		if !ok {
			log.Warningf("Element in group list is not string: %v, list: %v", e, g)
			continue
		}
		res = append(res, s)
	}
	return res, true
}

type populate func(groupNames []string) ([]int, error)

func populateGroupsDB(groupNames []string) ([]int, error) {
	ctx := orm.Context()
	cfg, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("failed to get OIDC config, error: %v", err)
		return nil, err
	}
	log.Debugf("populateGroupsDB, group filter %v", cfg.GroupFilter)
	return usergroup.Mgr.Populate(orm.Context(), model.UserGroupsFromName(filterGroup(groupNames, cfg.GroupFilter), common.OIDCGroupType))
}

// filterGroup filter group with a regular expression filter
func filterGroup(groupNames []string, filter string) []string {
	if len(filter) == 0 {
		return groupNames
	}
	pattern, err := regexp.Compile(filter)
	if err != nil {
		log.Errorf("failed to filter group, invalid filter %v", filter)
		return groupNames
	}
	result := make([]string, 0)
	for _, name := range groupNames {
		if pattern.MatchString(name) {
			result = append(result, name)
		}
	}
	log.Debugf("filter is %v, result is %v", filter, result)
	return result
}

// InjectGroupsToUser populates the group to DB and inject the group IDs to user model.
// The third optional parm is for UT only.
func InjectGroupsToUser(info *UserInfo, user *models.User, f ...populate) {
	if info == nil || user == nil {
		log.Warningf("user info or user model is nil, skip the func")
		return
	}
	var populateGroups populate
	if len(f) == 0 {
		populateGroups = populateGroupsDB
	} else {
		populateGroups = f[0]
	}
	if gids, err := populateGroups(info.Groups); err != nil {
		log.Warningf("failed to get group ID, error: %v, skip populating groups", err)
	} else {
		user.GroupIDs = gids
	}
	user.AdminRoleInAuth = info.AdminGroupMember
}

// Conn wraps connection info of an OIDC endpoint
type Conn struct {
	URL        string `json:"url"`
	VerifyCert bool   `json:"verify_cert"`
}

// TestEndpoint tests whether the endpoint is a valid OIDC endpoint.
// The nil return value indicates the success of the test
func TestEndpoint(conn Conn) error {
	// gooidc will try to call the discovery api when creating the provider and that's all we need to check
	ctx := clientCtx(context.Background(), conn.VerifyCert)
	_, err := gooidc.NewProvider(ctx, conn.URL)
	return err
}
