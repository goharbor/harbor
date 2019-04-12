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

// Package ng ...
// TODO rename the package name after removing ng
package ng

import (
	"time"

	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/utils/log"
	cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/replication/ng/config"
	"github.com/goharbor/harbor/src/replication/ng/event"
	"github.com/goharbor/harbor/src/replication/ng/operation"
	"github.com/goharbor/harbor/src/replication/ng/policy"
	"github.com/goharbor/harbor/src/replication/ng/policy/controller"
	"github.com/goharbor/harbor/src/replication/ng/registry"

	// register the Harbor adapter
	_ "github.com/goharbor/harbor/src/replication/ng/adapter/harbor"
)

var (
	// PolicyCtl is a global policy controller
	PolicyCtl policy.Controller
	// RegistryMgr is a global registry manager
	RegistryMgr registry.Manager
	// OperationCtl is a global operation controller
	OperationCtl operation.Controller
	// EventHandler handles images/chart pull/push events
	EventHandler event.Handler
)

// Init the global variables and configurations
func Init(closing chan struct{}) error {
	// init config
	registryURL, err := cfg.RegistryURL()
	if err != nil {
		return err
	}
	secretKey, err := cfg.SecretKey()
	if err != nil {
		return err
	}
	config.Config = &config.Configuration{
		CoreURL:          cfg.InternalCoreURL(),
		RegistryURL:      registryURL,
		TokenServiceURL:  cfg.InternalTokenServiceEndpoint(),
		JobserviceURL:    cfg.InternalJobServiceURL(),
		SecretKey:        secretKey,
		CoreSecret:       cfg.CoreSecret(),
		JobserviceSecret: cfg.JobserviceSecret(),
	}
	// TODO use a global http transport
	js := job.NewDefaultClient(config.Config.JobserviceURL, config.Config.CoreSecret)
	// init registry manager
	RegistryMgr = registry.NewDefaultManager()
	// init policy controller
	PolicyCtl = controller.NewController(js)
	// init operation controller
	OperationCtl = operation.NewController(js)
	// init event handler
	EventHandler = event.NewHandler(PolicyCtl, RegistryMgr, OperationCtl)
	log.Debug("the replication initialization completed")

	// Start health checker for registries
	go registry.NewHealthChecker(time.Minute*5, closing).Run()
	return nil
}
