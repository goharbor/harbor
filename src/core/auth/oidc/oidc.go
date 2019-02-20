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
	"net/http"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/token"
	"github.com/goharbor/harbor/src/common/utils/log"

	"github.com/goharbor/harbor/src/core/config"

	oidc "github.com/coreos/go-oidc"
	"github.com/coreos/go-oidc/oauth2"
)

// OauthClient abstracts the authorization code flow of an OAuth client
type OauthClient interface {
	AuthCodeURL(state string) (string, error)
	RequestToken(code string) (*models.User, error)
}

// defaultOauthClient is the implementation of OauthClient to use OIDC
type defaultOauthClient struct {
	OIDCConfig *oidc.Config
	provider   *oidc.Provider
	client     *oauth2.Client
}

// AuthCodeURL returns a URL to which the user can be redirected to start the
// OAuth authorization code flow
func (c *defaultOauthClient) AuthCodeURL(state string) (string, error) {
	client, err := c.ensureClient()
	if err != nil {
		return "", err
	}

	return client.AuthCodeURL(state, oauth2.GrantTypeAuthCode, ""), nil
}

// RequestToken retrieves a token from the OIDC provider using the provided code
func (c *defaultOauthClient) RequestToken(code string) (*models.User, error) {
	_, err := c.ensureClient()
	if err != nil {
		return nil, err
	}

	resp, err := c.client.RequestToken(oauth2.GrantTypeAuthCode, code)
	if err != nil {
		return nil, err
	}
	log.Infof("%s", resp.IDToken)

	v := c.provider.Verifier(c.OIDCConfig)
	idt, err := v.Verify(context.Background(), resp.IDToken)
	if err != nil {
		return nil, err
	}

	claims := &token.UserClaims{}
	err = idt.Claims(claims)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		Username: claims.Username,
		Email:    claims.Email,
	}

	return u, nil
}

func (c *defaultOauthClient) ensureClient() (*oauth2.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	endpoint, err := config.ExtEndpoint()
	if err != nil {
		return nil, err
	}

	providerURL, err := config.OIDCProvider()
	if err != nil {
		return nil, err
	}
	clientID, err := config.OIDCClientID()
	if err != nil {
		return nil, err
	}
	clientSecret, err := config.OIDCClientSecret()
	if err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(context.Background(), providerURL)
	if err != nil {
		return nil, err
	}

	hc := &http.Client{}
	client, err := oauth2.NewClient(hc, oauth2.Config{
		Credentials: oauth2.ClientCredentials{
			ID:     clientID,
			Secret: clientSecret,
		},
		AuthURL:     provider.Endpoint().AuthURL,
		TokenURL:    provider.Endpoint().TokenURL,
		RedirectURL: endpoint + "/c/oauth2/callback",
		Scope:       []string{"openid", "profile", "email", "groups"},
		AuthMethod:  oauth2.AuthMethodClientSecretBasic,
	})
	if err != nil {
		return nil, err
	}

	c.client = client

	return c.client, nil
}

// Client configures and returns an instance of defaultOauthClient
func Client() (OauthClient, error) {
	harborURL, err := config.ExtEndpoint()
	if err != nil {
		return nil, err
	}

	providerURL, err := config.OIDCProvider()
	if err != nil {
		return nil, err
	}
	clientID, err := config.OIDCClientID()
	if err != nil {
		return nil, err
	}
	clientSecret, err := config.OIDCClientSecret()
	if err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(context.Background(), providerURL)
	if err != nil {
		return nil, err
	}

	hc := &http.Client{}
	client, err := oauth2.NewClient(hc, oauth2.Config{
		Credentials: oauth2.ClientCredentials{
			ID:     clientID,
			Secret: clientSecret,
		},
		AuthURL:     provider.Endpoint().AuthURL,
		TokenURL:    provider.Endpoint().TokenURL,
		RedirectURL: harborURL + "/c/oauth2/callback",
		Scope:       []string{"openid", "profile", "email", "groups"},
		AuthMethod:  oauth2.AuthMethodClientSecretBasic,
	})
	if err != nil {
		return nil, err
	}
	return &defaultOauthClient{
		OIDCConfig: &oidc.Config{
			ClientID: clientID,
		},
		client:   client,
		provider: provider,
	}, nil
}
