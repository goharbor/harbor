package docker

import (
	"context"
	"net/http"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	ctypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/docker/registry"
)

// Resolver returns a new authenticated resolver.
func (c *Client) Resolver(_ context.Context, client *http.Client, plainHTTP bool) (remotes.Resolver, error) {
	return docker.NewResolver(docker.ResolverOptions{
		Credentials: c.Credential,
		Client:      client,
		PlainHTTP:   plainHTTP,
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
