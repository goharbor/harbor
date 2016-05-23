/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
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

package service

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"

	"github.com/astaxie/beego"
)

// NotificationHandler handles request on /service/notifications/, which listens to registry's events.
type NotificationHandler struct {
	beego.Controller
}

const manifestPattern = `^application/vnd.docker.distribution.manifest.v\d\+json`

// Post handles POST request, and records audit log or refreshes cache based on event.
func (n *NotificationHandler) Post() {
	var notification models.Notification
	//log.Info("Notification Handler triggered!\n")
	//	log.Infof("request body in string: %s", string(n.Ctx.Input.CopyBody()))
	err := json.Unmarshal(n.Ctx.Input.CopyBody(1<<32), &notification)

	if err != nil {
		log.Errorf("error while decoding json: %v", err)
		return
	}
	var username, action, repo, project, repoTag, url string
	var timestamp time.Time
	var matched bool
	for _, e := range notification.Events {
		matched, err = regexp.MatchString(manifestPattern, e.Target.MediaType)
		if err != nil {
			log.Errorf("Failed to match the media type against pattern, error: %v", err)
			matched = false
		}
		if matched && strings.HasPrefix(e.Request.UserAgent, "docker") {
			timestamp = e.TimeStamp
			url = e.Target.URL
			username = e.Actor.Name
			action = e.Action
			repo = e.Target.Repository
			repoTag = e.Target.Tag
			log.Debugf("repo tag is : %v ", repoTag)

			if strings.Contains(repo, "/") {
				project = repo[0:strings.LastIndex(repo, "/")]
			}
			if username == "" {
				username = "anonymous"
			}
			go dao.AccessLog(username, project, repo, repoTag, action)
			if action == "push" {
				go func() {
					err2 := svc_utils.RefreshCatalogCache()
					if err2 != nil {
						log.Errorf("Error happens when refreshing cache: %v", err2)
					}
				}()
				go func() {
					exist, err4ReopExists := dao.RepoExists(repo)
					if err4ReopExists != nil {
						log.Errorf("Error happened checking repo existence in db, error: %v, repo name: %s", err4ReopExists, repo)
					}
					if exist {
						return
					}
					log.Debugf("Add repo %s into DB.", repo)
					repoItem := models.RepoRecord{Name: repo, OwnerName: username, ProjectName: project, Created: timestamp, URL: url}
					_, err4AddRepo := dao.AddRepo(repoItem)
					if err4AddRepo != nil {
						log.Errorf("Error happens when adding repo: %v", err4AddRepo)
					}
				}()

			}
			if action == "pull" {
				go func() {
					log.Debugf("Increase the repo %s pull count.", repo)
					repoItem := models.RepoRecord{Name: repo}
					err4IncreasePullCount := dao.IncreasePullCount(repoItem)
					if err4IncreasePullCount != nil {
						log.Errorf("Error happens when increaing pull count: %v", err4IncreasePullCount)
					}
				}()
			}
		}
	}

}

// Render returns nil as it won't render any template.
func (n *NotificationHandler) Render() error {
	return nil
}
