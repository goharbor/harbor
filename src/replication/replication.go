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

package replication

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/controller/replication"
	cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/policy"
	"github.com/goharbor/harbor/src/replication/policy/controller"
	"github.com/goharbor/harbor/src/replication/registry"

	// register the Harbor adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/harbor"
	// register the DockerHub adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/dockerhub"
	// register the Native adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/native"
	// register the huawei adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/huawei"
	// register the Google Gcr adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/googlegcr"
	// register the AwsEcr adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/awsecr"
	// register the AzureAcr adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/azurecr"
	// register the AliACR adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/aliacr"
	// register the Jfrog Artifactory adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/jfrog"
	// register the Quay.io adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/quay"
	// register the Helm Hub adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/helmhub"
	// register the GitLab adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/gitlab"
	// register the DTR adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/dtr"
	// register the Artifact Hub adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/artifacthub"
)

var (
	// PolicyCtl is a global policy controller
	PolicyCtl policy.Controller
	// RegistryMgr is a global registry manager
	RegistryMgr registry.Manager
	// EventHandler handles images/chart pull/push events
	EventHandler event.Handler
)

func init() {
	callbackFunc := func(ctx context.Context, param string) error {
		policyID, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return err
		}

		policy, err := PolicyCtl.Get(policyID)
		if err != nil {
			return err
		}
		if policy == nil {
			return fmt.Errorf("policy %d not found", policyID)
		}
		if err = event.PopulateRegistries(RegistryMgr, policy); err != nil {
			return err
		}

		_, err = replication.Ctl.Start(ctx, policy, nil, task.ExecutionTriggerSchedule)
		return err
	}
	err := scheduler.RegisterCallbackFunc(controller.CallbackFuncName, callbackFunc)
	if err != nil {
		log.Errorf("failed to register the callback function for replication: %v", err)
	}
}

// Init the global variables and configurations
func Init(closing, done chan struct{}) error {
	// init config
	secretKey, err := cfg.SecretKey()
	if err != nil {
		return err
	}
	config.Config = &config.Configuration{
		CoreURL:          cfg.InternalCoreURL(),
		TokenServiceURL:  cfg.InternalTokenServiceEndpoint(),
		SecretKey:        secretKey,
		JobserviceSecret: cfg.JobserviceSecret(),
	}
	// init registry manager
	RegistryMgr = registry.NewDefaultManager()
	// init policy controller
	PolicyCtl = controller.NewController()
	// init event handler
	EventHandler = event.NewHandler(PolicyCtl, RegistryMgr)
	log.Debug("the replication initialization completed")

	// Start health checker for registries
	go registry.NewHealthChecker(time.Minute*5, closing, done).Run()
	return nil
}
