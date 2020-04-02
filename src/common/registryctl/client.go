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

package registryctl

import (
	"os"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/registryctl/client"
)

var (
	// RegistryCtlClient is a client for registry controller
	RegistryCtlClient client.Client
)

// Init ...
func Init() {
	initRegistryCtlClient()
}

func initRegistryCtlClient() {
	registryCtlURL := os.Getenv("REGISTRY_CONTROLLER_URL")
	if len(registryCtlURL) == 0 {
		registryCtlURL = common.DefaultRegistryControllerEndpoint
	}

	log.Infof("initializing client for registry %s ...", registryCtlURL)
	cfg := &client.Config{
		Secret: os.Getenv("JOBSERVICE_SECRET"),
	}
	RegistryCtlClient = client.NewClient(registryCtlURL, cfg)
}
