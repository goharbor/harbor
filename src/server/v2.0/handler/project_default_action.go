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
package handler

import (
	"strings"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
)

const (
	defaultWebHookGlobalCreator = "system"
)

//createDefaultNotifyPolicy if enabel global webhook, will be added to all new projects
func (a *projectAPI) createDefaultNotifyPolicy(projectID int64) error {
	webHookGlobalConfig := config.GlobalWebHook()
	if webHookGlobalConfig.Enable {
		policy := &commonmodels.NotificationPolicy{
			Name:        webHookGlobalConfig.Name,
			Description: webHookGlobalConfig.Description,
			ProjectID:   projectID,
			Targets: []commonmodels.EventTarget{
				{
					Type:           webHookGlobalConfig.TargetType,
					Address:        webHookGlobalConfig.TargetAddress,
					AuthHeader:     webHookGlobalConfig.TargetAuthHeader,
					SkipCertVerify: webHookGlobalConfig.TargetSkipCertVerify,
				},
			},
			EventTypes: strings.Split(webHookGlobalConfig.EventTypes, ","),
			Creator:    defaultWebHookGlobalCreator,
			Enabled:    true,
		}
		id, err := notification.PolicyMgr.Create(policy)
		if err != nil {
			return err
		}
		log.Infof("project %d create default webHookPolicy %d", projectID, id)
		return nil
	}
	return nil
}
