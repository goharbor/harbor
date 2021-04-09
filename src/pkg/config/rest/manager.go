//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package rest

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	libCfg "github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/config"
	"github.com/goharbor/harbor/src/pkg/config/store"
	"os"
	"strings"
)

func init() {
	unitTest := os.Getenv("UTTEST")
	if strings.EqualFold(unitTest, "true") == true {
		libCfg.Register(common.RestCfgManager, NewRESTCfgManager("sample_url", "sample_secret"))
		return
	}

	coreURL := os.Getenv("CORE_URL")
	if len(coreURL) == 0 {
		return
	}
	configURL := coreURL + common.CoreConfigPath
	jobSvcSecret := os.Getenv("JOBSERVICE_SECRET")
	if len(jobSvcSecret) == 0 {
		return
	}
	// by default rest config manager is used by jobservice
	// for other scenario, should change the initialization of config manager
	libCfg.Register(common.RestCfgManager, NewRESTCfgManager(configURL, jobSvcSecret))
}

// NewRESTCfgManager - create REST config manager
func NewRESTCfgManager(configURL, secret string) *config.CfgManager {
	secAuth := auth.NewSecretAuthorizer(secret)
	manager := &config.CfgManager{Store: store.NewConfigStore(NewRESTDriver(configURL, secAuth))}
	return manager
}
