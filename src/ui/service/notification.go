// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package service

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/api"

	"github.com/astaxie/beego"
)

// NotificationHandler handles request on /service/notifications/, which listens to registry's events.
type NotificationHandler struct {
	beego.Controller
}

const manifestPattern = `^application/vnd.docker.distribution.manifest.v\d\+(json|prettyjws)`
const vicPrefix = "vic/"

// Post handles POST request, and records audit log or refreshes cache based on event.
func (n *NotificationHandler) Post() {
	var notification models.Notification
	err := json.Unmarshal(n.Ctx.Input.CopyBody(1<<32), &notification)

	if err != nil {
		log.Errorf("failed to decode notification: %v", err)
		return
	}

	events, err := filterEvents(&notification)
	if err != nil {
		log.Errorf("failed to filter events: %v", err)
		return
	}

	for _, event := range events {
		repository := event.Target.Repository

		project, _ := utils.ParseRepository(repository)
		tag := event.Target.Tag
		action := event.Action

		user := event.Actor.Name
		if len(user) == 0 {
			user = "anonymous"
		}

		go func() {
			if err := dao.AccessLog(user, project, repository, tag, action); err != nil {
				log.Errorf("failed to add access log: %v", err)
			}
		}()
		if action == "push" {
			go func() {
				exist := dao.RepositoryExists(repository)
				if exist {
					return
				}
				log.Debugf("Add repository %s into DB.", repository)
				repoRecord := models.RepoRecord{Name: repository, OwnerName: user, ProjectName: project}
				if err := dao.AddRepository(repoRecord); err != nil {
					log.Errorf("Error happens when adding repository: %v", err)
				}
			}()
			go api.TriggerReplicationByRepository(repository, []string{tag}, models.RepOpTransfer)
		}
		if action == "pull" {
			go func() {
				log.Debugf("Increase the repository %s pull count.", repository)
				if err := dao.IncreasePullCount(repository); err != nil {
					log.Errorf("Error happens when increasing pull count: %v", repository)
				}
			}()
		}
	}
}

func filterEvents(notification *models.Notification) ([]*models.Event, error) {
	events := []*models.Event{}

	for _, event := range notification.Events {
		log.Debugf("receive an event: \n----ID: %s \n----target: %s:%s \n----digest: %s \n----action: %s \n----mediatype: %s \n----user-agent: %s", event.ID, event.Target.Repository,
			event.Target.Tag, event.Target.Digest, event.Action, event.Target.MediaType, event.Request.UserAgent)

		isManifest, err := regexp.MatchString(manifestPattern, event.Target.MediaType)
		if err != nil {
			log.Errorf("failed to match the media type against pattern: %v", err)
			continue
		}

		if !isManifest {
			continue
		}

		//pull and push manifest by docker-client or vic
		if (strings.HasPrefix(event.Request.UserAgent, "docker") || strings.HasPrefix(event.Request.UserAgent, vicPrefix)) &&
			(event.Action == "pull" || event.Action == "push") {
			events = append(events, &event)
			log.Debugf("add event to collect: %s", event.ID)
			continue
		}

		//push manifest by docker-client or job-service
		if strings.ToLower(strings.TrimSpace(event.Request.UserAgent)) == "harbor-registry-client" && event.Action == "push" {
			events = append(events, &event)
			log.Debugf("add event to collect: %s", event.ID)
			continue
		}
	}

	return events, nil
}

// Render returns nil as it won't render any template.
func (n *NotificationHandler) Render() error {
	return nil
}
