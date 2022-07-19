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

	"github.com/docker/cli/cli/config/configfile"

	"oras.land/oras-go/pkg/auth"
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
