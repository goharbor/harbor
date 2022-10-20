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

package docker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/registry"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// The following functions are adapted from github.com/docker/docker/registry
// We need these to support passing in a transport that has custom TLS configuration
// They are not exposed in the docker/registry package that's why they are copied here

type loginCredentialStore struct {
	authConfig *types.AuthConfig
}

func (lcs loginCredentialStore) Basic(*url.URL) (string, string) {
	return lcs.authConfig.Username, lcs.authConfig.Password
}

func (lcs loginCredentialStore) RefreshToken(*url.URL, string) string {
	return lcs.authConfig.IdentityToken
}

func (lcs loginCredentialStore) SetRefreshToken(u *url.URL, service, token string) {
	lcs.authConfig.IdentityToken = token
}

// loginWithTLS tries to login to the v2 registry server.
// A custom tls.Config is used to override the default TLS configuration of the different registry endpoints.
// The tls.Config is created using the provided certificate, certificate key and certificate authority.
func (c *Client) loginWithTLS(ctx context.Context, service registry.Service, certFile, keyFile, caFile string, authConfig *types.AuthConfig, userAgent string) (string, string, error) {
	tlsConfig, err := tlsconfig.Client(tlsconfig.Options{CAFile: caFile, CertFile: certFile, KeyFile: keyFile})
	if err != nil {
		return "", "", err
	}

	endpoints, err := c.getEndpoints(authConfig.ServerAddress, service)
	if err != nil {
		return "", "", err
	}

	var status, token string
	for _, endpoint := range endpoints {
		endpoint.TLSConfig = tlsConfig
		status, token, err = loginV2(authConfig, endpoint, userAgent)

		if err != nil {
			if isNotAuthorizedError(err) {
				return "", "", err
			}

			logrus.WithError(err).Infof("Error logging in to endpoint, trying next endpoint")
			continue
		}

		return status, token, nil
	}

	return "", "", err
}

// getEndpoints returns the endpoints for the given hostname.
func (c *Client) getEndpoints(address string, service registry.Service) ([]registry.APIEndpoint, error) {
	var registryHostName = IndexHostname

	if address != "" {
		if !strings.HasPrefix(address, "https://") && !strings.HasPrefix(address, "http://") {
			address = fmt.Sprintf("https://%s", address)
		}
		u, err := url.Parse(address)
		if err != nil {
			return nil, errdefs.InvalidParameter(errors.Wrapf(err, "unable to parse server address"))
		}
		registryHostName = u.Host
	}

	// Lookup endpoints for authentication using "LookupPushEndpoints", which
	// excludes mirrors to prevent sending credentials of the upstream registry
	// to a mirror.
	endpoints, err := service.LookupPushEndpoints(registryHostName)
	if err != nil {
		return nil, errdefs.InvalidParameter(err)
	}

	return endpoints, nil
}

// loginV2 tries to login to the v2 registry server. The given registry
// endpoint will be pinged to get authorization challenges. These challenges
// will be used to authenticate against the registry to validate credentials.
func loginV2(authConfig *types.AuthConfig, endpoint registry.APIEndpoint, userAgent string) (string, string, error) {
	var (
		endpointStr          = strings.TrimRight(endpoint.URL.String(), "/") + "/v2/"
		modifiers            = registry.Headers(userAgent, nil)
		authTransport        = transport.NewTransport(newTransport(endpoint.TLSConfig), modifiers...)
		credentialAuthConfig = *authConfig
		creds                = loginCredentialStore{authConfig: &credentialAuthConfig}
	)

	logrus.Debugf("attempting v2 login to registry endpoint %s", endpointStr)

	loginClient, err := v2AuthHTTPClient(endpoint.URL, authTransport, modifiers, creds, nil)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodGet, endpointStr, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := loginClient.Do(req)
	if err != nil {
		err = translateV2AuthError(err)
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "Login Succeeded", credentialAuthConfig.IdentityToken, nil
	}

	// TODO(dmcgowan): Attempt to further interpret result, status code and error code string
	return "", "", errors.Errorf("login attempt to %s failed with status: %d %s", endpointStr, resp.StatusCode, http.StatusText(resp.StatusCode))
}

// newTransport returns a new HTTP transport. If tlsConfig is nil, it uses the
// default TLS configuration.
func newTransport(tlsConfig *tls.Config) *http.Transport {
	if tlsConfig == nil {
		tlsConfig = tlsconfig.ServerDefault()
	}

	direct := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         direct.DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
		// TODO(dmcgowan): Call close idle connections when complete and use keep alive
		DisableKeepAlives: true,
	}
}

func v2AuthHTTPClient(endpoint *url.URL, authTransport http.RoundTripper, modifiers []transport.RequestModifier, creds auth.CredentialStore, scopes []auth.Scope) (*http.Client, error) {
	challengeManager, _, err := registry.PingV2Registry(endpoint, authTransport)
	if err != nil {
		return nil, err
	}

	tokenHandlerOptions := auth.TokenHandlerOptions{
		Transport:     authTransport,
		Credentials:   creds,
		OfflineAccess: true,
		ClientID:      registry.AuthClientID,
		Scopes:        scopes,
	}
	tokenHandler := auth.NewTokenHandlerWithOptions(tokenHandlerOptions)
	basicHandler := auth.NewBasicHandler(creds)
	modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler))

	return &http.Client{
		Transport: transport.NewTransport(authTransport, modifiers...),
		Timeout:   15 * time.Second,
	}, nil
}

func translateV2AuthError(err error) error {
	switch e := err.(type) {
	case *url.Error:
		switch e2 := e.Err.(type) {
		case errcode.Error:
			switch e2.Code {
			case errcode.ErrorCodeUnauthorized:
				return errdefs.Unauthorized(err)
			}
		}
	}

	return err
}

func isNotAuthorizedError(err error) bool {
	return strings.Contains(err.Error(), "401 Unauthorized")
}
