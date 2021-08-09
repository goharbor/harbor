package docker

import (
	"context"

	"github.com/deislabs/oras/pkg/auth"

	"github.com/docker/cli/cli/config/configfile"
)

// Logout logs out from a docker registry identified by the hostname.
func (c *Client) Logout(_ context.Context, hostname string) error {
	hostname = resolveHostname(hostname)

	var configs []*configfile.ConfigFile
	for _, config := range c.configs {
		if _, ok := config.AuthConfigs[hostname]; ok {
			configs = append(configs, config)
		}
	}
	if len(configs) == 0 {
		return auth.ErrNotLoggedIn
	}

	// Log out form the primary config only as backups are read-only.
	return c.primaryCredentialsStore(hostname).Erase(hostname)
}
