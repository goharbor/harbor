/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"oras.land/oras-go/pkg/registry/remote/internal/errutil"
)

// DefaultClient is the default auth-decorated client.
var DefaultClient = &Client{
	Header: http.Header{
		"User-Agent": {"oras-go"},
	},
	Cache: DefaultCache,
}

// maxResponseBytes specifies the default limit on how many response bytes are
// allowed in the server's response from authorization service servers.
// A typical response message from authorization service servers is around 1 to
// 4 KiB. Since the size of a token must be smaller than the HTTP header size
// limit, which is usually 16 KiB. As specified by the distribution, the
// response may contain 2 identical tokens, that is, 16 x 2 = 32 KiB.
// Hence, 128 KiB should be sufficient.
// References: https://docs.docker.com/registry/spec/auth/token/
var maxResponseBytes int64 = 128 * 1024 // 128 KiB

// defaultClientID specifies the default client ID used in OAuth2.
// See also ClientID.
var defaultClientID = "oras-go"

// Client is an auth-decorated HTTP client.
// Its zero value is a usable client that uses http.DefaultClient with no cache.
type Client struct {
	// Client is the underlying HTTP client used to access the remote
	// server.
	// If nil, http.DefaultClient is used.
	Client *http.Client

	// Header contains the custom headers to be added to each request.
	Header http.Header

	// Credential specifies the function for resolving the credential for the
	// given registry (i.e. host:port).
	// `EmptyCredential` is a valid return value and should not be considered as
	// an error.
	// If nil, the credential is always resolved to `EmptyCredential`.
	Credential func(context.Context, string) (Credential, error)

	// Cache caches credentials for direct accessing the remote registry.
	// If nil, no cache is used.
	Cache Cache

	// ClientID used in fetching OAuth2 token as a required field.
	// If empty, a default client ID is used.
	// Reference: https://docs.docker.com/registry/spec/auth/oauth/#getting-a-token
	ClientID string

	// ForceAttemptOAuth2 controls whether to follow OAuth2 with password grant
	// instead the distribution spec when authenticating using username and
	// password.
	// References:
	// - https://docs.docker.com/registry/spec/auth/jwt/
	// - https://docs.docker.com/registry/spec/auth/oauth/
	ForceAttemptOAuth2 bool
}

// client returns an HTTP client used to access the remote registry.
// http.DefaultClient is return if the client is not configured.
func (c *Client) client() *http.Client {
	if c.Client == nil {
		return http.DefaultClient
	}
	return c.Client
}

// send adds headers to the request and sends the request to the remote server.
func (c *Client) send(req *http.Request) (*http.Response, error) {
	for key, values := range c.Header {
		req.Header[key] = append(req.Header[key], values...)
	}
	return c.client().Do(req)
}

// credential resolves the credential for the given registry.
func (c *Client) credential(ctx context.Context, reg string) (Credential, error) {
	if c.Credential == nil {
		return EmptyCredential, nil
	}
	return c.Credential(ctx, reg)
}

// cache resolves the cache.
// noCache is return if the cache is not configured.
func (c *Client) cache() Cache {
	if c.Cache == nil {
		return noCache{}
	}
	return c.Cache
}

// SetUserAgent sets the user agent for all out-going requests.
func (c *Client) SetUserAgent(userAgent string) {
	if c.Header == nil {
		c.Header = http.Header{}
	}
	c.Header.Set("User-Agent", userAgent)
}

// Do sends the request to the remote server with resolving authentication
// attempted.
// On authentication failure due to bad credential,
// - Do returns error if it fails to fetch token for bearer auth.
// - Do returns the registry response without error for basic auth.
func (c *Client) Do(originalReq *http.Request) (*http.Response, error) {
	ctx := originalReq.Context()
	req := originalReq.Clone(ctx)

	// attempt cached auth token
	var attemptedKey string
	cache := c.cache()
	registry := originalReq.Host
	scheme, err := cache.GetScheme(ctx, registry)
	if err == nil {
		switch scheme {
		case SchemeBasic:
			token, err := cache.GetToken(ctx, registry, SchemeBasic, "")
			if err == nil {
				req.Header.Set("Authorization", "Basic "+token)
			}
		case SchemeBearer:
			scopes := GetScopes(ctx)
			attemptedKey = strings.Join(scopes, " ")
			token, err := cache.GetToken(ctx, registry, SchemeBearer, attemptedKey)
			if err == nil {
				req.Header.Set("Authorization", "Bearer "+token)
			}
		}
	}

	resp, err := c.send(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	// attempt again with credentials for recognized schemes
	challenge := resp.Header.Get("Www-Authenticate")
	scheme, params := parseChallenge(challenge)
	switch scheme {
	case SchemeBasic:
		resp.Body.Close()

		token, err := cache.Set(ctx, registry, SchemeBasic, "", func(ctx context.Context) (string, error) {
			return c.fetchBasicAuth(ctx, registry)
		})
		if err != nil {
			return nil, fmt.Errorf("%s %q: %w", resp.Request.Method, resp.Request.URL, err)
		}

		req = originalReq.Clone(ctx)
		req.Header.Set("Authorization", "Basic "+token)
	case SchemeBearer:
		resp.Body.Close()

		// merge hinted scopes with challenged scopes
		scopes := GetScopes(ctx)
		if scope := params["scope"]; scope != "" {
			scopes = append(scopes, strings.Split(scope, " ")...)
			scopes = CleanScopes(scopes)
		}
		key := strings.Join(scopes, " ")

		// attempt the cache again if there is a scope change
		if key != attemptedKey {
			if token, err := cache.GetToken(ctx, registry, SchemeBearer, key); err == nil {
				req = originalReq.Clone(ctx)
				req.Header.Set("Authorization", "Bearer "+token)

				resp, err := c.send(req)
				if err != nil {
					return nil, err
				}
				if resp.StatusCode != http.StatusUnauthorized {
					return resp, nil
				}
				resp.Body.Close()
			}
		}

		// attempt with credentials
		realm := params["realm"]
		service := params["service"]
		token, err := cache.Set(ctx, registry, SchemeBearer, key, func(ctx context.Context) (string, error) {
			return c.fetchBearerToken(ctx, registry, realm, service, scopes)
		})
		if err != nil {
			return nil, fmt.Errorf("%s %q: %w", resp.Request.Method, resp.Request.URL, err)
		}

		req = originalReq.Clone(ctx)
		req.Header.Set("Authorization", "Bearer "+token)
	default:
		return resp, nil
	}

	return c.send(req)
}

// fetchBasicAuth fetches a basic auth token for the basic challenge.
func (c *Client) fetchBasicAuth(ctx context.Context, registry string) (string, error) {
	cred, err := c.credential(ctx, registry)
	if err != nil {
		return "", fmt.Errorf("failed to resolve credential: %w", err)
	}
	if cred == EmptyCredential {
		return "", errors.New("credential required for basic auth")
	}
	if cred.Username == "" || cred.Password == "" {
		return "", errors.New("missing username or password for basic auth")
	}
	auth := cred.Username + ":" + cred.Password
	return base64.StdEncoding.EncodeToString([]byte(auth)), nil
}

// fetchBearerToken fetches an access token for the bearer challenge.
func (c *Client) fetchBearerToken(ctx context.Context, registry, realm, service string, scopes []string) (string, error) {
	cred, err := c.credential(ctx, registry)
	if err != nil {
		return "", err
	}
	if cred.AccessToken != "" {
		return cred.AccessToken, nil
	}
	if cred == EmptyCredential || (cred.RefreshToken == "" && !c.ForceAttemptOAuth2) {
		return c.fetchDistributionToken(ctx, realm, service, scopes, cred.Username, cred.Password)
	}
	return c.fetchOAuth2Token(ctx, realm, service, scopes, cred)
}

// fetchDistributionToken fetches an access token as defined by the distribution
// specification.
// It fetches anonymous tokens if no credential is provided.
// References:
// - https://docs.docker.com/registry/spec/auth/jwt/
// - https://docs.docker.com/registry/spec/auth/token/
func (c *Client) fetchDistributionToken(ctx context.Context, realm, service string, scopes []string, username, password string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, realm, nil)
	if err != nil {
		return "", err
	}
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}
	q := req.URL.Query()
	if service != "" {
		q.Set("service", service)
	}
	for _, scope := range scopes {
		q.Add("scope", scope)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.send(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errutil.ParseErrorResponse(resp)
	}

	// As specified in https://docs.docker.com/registry/spec/auth/token/ section
	// "Token Response Fields", the token is either in `token` or
	// `access_token`. If both present, they are identical.
	var result struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	lr := io.LimitReader(resp.Body, maxResponseBytes)
	if err := json.NewDecoder(lr).Decode(&result); err != nil {
		return "", fmt.Errorf("%s %q: failed to decode response: %w", resp.Request.Method, resp.Request.URL, err)
	}
	if result.AccessToken != "" {
		return result.AccessToken, nil
	}
	if result.Token != "" {
		return result.Token, nil
	}
	return "", fmt.Errorf("%s %q: empty token returned", resp.Request.Method, resp.Request.URL)
}

// fetchOAuth2Token fetches an OAuth2 access token.
// Reference: https://docs.docker.com/registry/spec/auth/oauth/
func (c *Client) fetchOAuth2Token(ctx context.Context, realm, service string, scopes []string, cred Credential) (string, error) {
	form := url.Values{}
	if cred.RefreshToken != "" {
		form.Set("grant_type", "refresh_token")
		form.Set("refresh_token", cred.RefreshToken)
	} else if cred.Username != "" && cred.Password != "" {
		form.Set("grant_type", "password")
		form.Set("username", cred.Username)
		form.Set("password", cred.Password)
	} else {
		return "", errors.New("missing username or password for bearer auth")
	}
	form.Set("service", service)
	clientID := c.ClientID
	if clientID == "" {
		clientID = defaultClientID
	}
	form.Set("client_id", clientID)
	if len(scopes) != 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}
	body := strings.NewReader(form.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, realm, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.send(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errutil.ParseErrorResponse(resp)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	lr := io.LimitReader(resp.Body, maxResponseBytes)
	if err := json.NewDecoder(lr).Decode(&result); err != nil {
		return "", fmt.Errorf("%s %q: failed to decode response: %w", resp.Request.Method, resp.Request.URL, err)
	}
	if result.AccessToken != "" {
		return result.AccessToken, nil
	}
	return "", fmt.Errorf("%s %q: empty token returned", resp.Request.Method, resp.Request.URL)
}
