// Copyright 2018 Project Harbor Authors
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
	"errors"
	"fmt"
	gooidc "github.com/coreos/go-oidc"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const googleEndpoint = "https://accounts.google.com"

type providerHelper struct {
	sync.Mutex
	instance     atomic.Value
	setting      atomic.Value
	creationTime time.Time
}

func (p *providerHelper) get() (*gooidc.Provider, error) {
	if p.instance.Load() != nil {
		if time.Now().Sub(p.creationTime) > 3*time.Second {
			if err := p.create(); err != nil {
				return nil, err
			}
		}
	} else {
		p.Lock()
		defer p.Unlock()
		if p.instance.Load() == nil {
			if err := p.reloadSetting(); err != nil {
				return nil, err
			}
			if err := p.create(); err != nil {
				return nil, err
			}
			go func() {
				for {
					if err := p.reloadSetting(); err != nil {
						log.Warningf("Failed to refresh configuration, error: %v", err)
					}
					time.Sleep(3 * time.Second)
				}
			}()
		}
	}

	return p.instance.Load().(*gooidc.Provider), nil
}

func (p *providerHelper) reloadSetting() error {
	conf, err := config.OIDCSetting()
	if err != nil {
		return fmt.Errorf("failed to load OIDC setting: %v", err)
	}
	p.setting.Store(*conf)
	return nil
}

func (p *providerHelper) create() error {
	if p.setting.Load() == nil {
		return errors.New("the configuration is not loaded")
	}
	s := p.setting.Load().(models.OIDCSetting)
	ctx := clientCtx(context.Background(), s.VerifyCert)
	provider, err := gooidc.NewProvider(ctx, s.Endpoint)
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
}

// Token wraps the attributes of a oauth2 token plus the attribute of ID token
type Token struct {
	oauth2.Token
	IDToken string `json:"id_token"`
}

func getOauthConf() (*oauth2.Config, error) {
	p, err := provider.get()
	if err != nil {
		return nil, err
	}
	setting := provider.setting.Load().(models.OIDCSetting)
	scopes := []string{}
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

// AuthCodeURL returns the URL for OIDC provider's consent page.  The state should be verified when user is redirected
// back to Harbor.
func AuthCodeURL(state string) (string, error) {
	conf, err := getOauthConf()
	if err != nil {
		log.Errorf("Failed to get OAuth configuration, error: %v", err)
		return "", err
	}
	if strings.HasPrefix(conf.Endpoint.AuthURL, googleEndpoint) { // make sure the refresh token will be returned
		return conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")), nil
	}
	return conf.AuthCodeURL(state), nil
}

// ExchangeToken get the token from token provider via the code
func ExchangeToken(ctx context.Context, code string) (*Token, error) {
	oauth, err := getOauthConf()
	if err != nil {
		log.Errorf("Failed to get OAuth configuration, error: %v", err)
		return nil, err
	}
	setting := provider.setting.Load().(models.OIDCSetting)
	ctx = clientCtx(ctx, setting.VerifyCert)
	oauthToken, err := oauth.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return &Token{Token: *oauthToken, IDToken: oauthToken.Extra("id_token").(string)}, nil
}

// VerifyToken verifies the ID token based on the OIDC settings
func VerifyToken(ctx context.Context, rawIDToken string) (*gooidc.IDToken, error) {
	p, err := provider.get()
	if err != nil {
		return nil, err
	}
	verifier := p.Verifier(&gooidc.Config{ClientID: provider.setting.Load().(models.OIDCSetting).ClientID})
	setting := provider.setting.Load().(models.OIDCSetting)
	ctx = clientCtx(ctx, setting.VerifyCert)
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

// RefreshToken refreshes the token passed in parameter, and return the new token.
func RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	oauth, err := getOauthConf()
	if err != nil {
		log.Errorf("Failed to get OAuth configuration, error: %v", err)
		return nil, err
	}
	setting := provider.setting.Load().(models.OIDCSetting)
	ctx = clientCtx(ctx, setting.VerifyCert)
	ts := oauth.TokenSource(ctx, &token.Token)
	t, err := ts.Token()
	if err != nil {
		return nil, err
	}
	it, ok := t.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("failed to get id_token from refresh response")
	}
	return &Token{Token: *t, IDToken: it}, nil
}
