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

	ctypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/registry"

	iface "oras.land/oras-go/pkg/auth"
)

const IndexHostname = "index.docker.io"

// Login logs in to a docker registry identified by the hostname.
// Deprecated: use LoginWithOpts
func (c *Client) Login(ctx context.Context, hostname, username, secret string, insecure bool) error {
	settings := &iface.LoginSettings{
		Context:  ctx,
		Hostname: hostname,
		Username: username,
		Secret:   secret,
		Insecure: insecure,
	}
	return c.login(settings)
}

// LoginWithOpts logs in to a docker registry identified by the hostname with custom options.
func (c *Client) LoginWithOpts(options ...iface.LoginOption) error {
	settings := &iface.LoginSettings{}
	for _, option := range options {
		option(settings)
	}
	return c.login(settings)
}

func (c *Client) login(settings *iface.LoginSettings) error {
	hostname := resolveHostname(settings.Hostname)
	cred := types.AuthConfig{
		Username:      settings.Username,
		ServerAddress: hostname,
	}
	if settings.Username == "" {
		cred.IdentityToken = settings.Secret
	} else {
		cred.Password = settings.Secret
	}

	opts := registry.ServiceOptions{}

	if settings.Insecure {
		opts.InsecureRegistries = []string{hostname}
	}

	// Login to ensure valid credential
	remote, err := registry.NewService(opts)
	if err != nil {
		return err
	}
	ctx := settings.Context
	if ctx == nil {
		ctx = context.Background()
	}
	userAgent := settings.UserAgent
	if userAgent == "" {
		userAgent = "oras"
	}

	var token string
	if (settings.CertFile != "" && settings.KeyFile != "") || settings.CAFile != "" {
		_, token, err = c.loginWithTLS(ctx, remote, settings.CertFile, settings.KeyFile, settings.CAFile, &cred, userAgent)
	} else {
		_, token, err = remote.Auth(ctx, &cred, userAgent)
	}

	if err != nil {
		return err
	}

	if token != "" {
		cred.Username = ""
		cred.Password = ""
		cred.IdentityToken = token
	}

	// Store credential
	return c.primaryCredentialsStore(hostname).Store(ctypes.AuthConfig(cred))
}
