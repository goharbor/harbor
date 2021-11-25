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
	"net/http"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	ctypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/docker/registry"

	iface "oras.land/oras-go/pkg/auth"
)

// Resolver returns a new authenticated resolver.
// Deprecated: use ResolverWithOpts
func (c *Client) Resolver(_ context.Context, client *http.Client, plainHTTP bool) (remotes.Resolver, error) {
	return docker.NewResolver(docker.ResolverOptions{
		Credentials: c.Credential,
		Client:      client,
		PlainHTTP:   plainHTTP,
	}), nil
}

// ResolverWithOpts returns a new authenticated resolver with custom options.
func (c *Client) ResolverWithOpts(options ...iface.ResolverOption) (remotes.Resolver, error) {
	settings := &iface.ResolverSettings{}
	for _, option := range options {
		option(settings)
	}
	return docker.NewResolver(docker.ResolverOptions{
		Credentials: c.Credential,
		Client:      settings.Client,
		PlainHTTP:   settings.PlainHTTP,
		Headers:     settings.Headers,
	}), nil
}

// Credential returns the login credential of the request host.
func (c *Client) Credential(hostname string) (string, string, error) {
	hostname = resolveHostname(hostname)
	var (
		auth ctypes.AuthConfig
		err  error
	)
	for _, cfg := range c.configs {
		auth, err = cfg.GetAuthConfig(hostname)
		if err != nil {
			// fall back to next config
			continue
		}
		if auth.IdentityToken != "" {
			return "", auth.IdentityToken, nil
		}
		if auth.Username == "" && auth.Password == "" {
			// fall back to next config
			continue
		}
		return auth.Username, auth.Password, nil
	}
	return "", "", err
}

// resolveHostname resolves Docker specific hostnames
func resolveHostname(hostname string) string {
	switch hostname {
	case registry.IndexHostname, registry.IndexName, registry.DefaultV2Registry.Host:
		return registry.IndexServer
	}
	return hostname
}
