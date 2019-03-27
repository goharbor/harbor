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
	"github.com/goharbor/harbor/src/common/utils/log"
	cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/replication/ng/config"
	"github.com/goharbor/harbor/src/replication/ng/operation"
	"github.com/goharbor/harbor/src/replication/ng/operation/execution"
	"github.com/goharbor/harbor/src/replication/ng/operation/scheduler"
	"github.com/goharbor/harbor/src/replication/ng/policy"
	"github.com/goharbor/harbor/src/replication/ng/registry"

	// register the Harbor adapter
	_ "github.com/goharbor/harbor/src/replication/ng/adapter/harbor"
)

var (
	// PolicyMgr is a global policy manager
	PolicyMgr policy.Manager
	// RegistryMgr is a global registry manager
	RegistryMgr registry.Manager
	// OperationCtl is a global operation controller
	OperationCtl operation.Controller
)

// Init the global variables and configurations
func Init() error {
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
		CoreURL:         cfg.InternalCoreURL(),
		RegistryURL:     registryURL,
		TokenServiceURL: cfg.InternalTokenServiceEndpoint(),
		JobserviceURL:   cfg.InternalJobServiceURL(),
		SecretKey:       secretKey,
		Secret:          cfg.CoreSecret(),
	}
	// Init registry manager
	RegistryMgr = registry.NewDefaultManager()
	// init policy manager
	PolicyMgr = policy.NewDefaultManager()
	// init operatoin controller
	OperationCtl = operation.NewController(execution.NewDefaultManager(), RegistryMgr,
		scheduler.NewScheduler(config.Config.JobserviceURL, config.Config.Secret))
	log.Debug("the replication initialization completed")
	return nil
}
