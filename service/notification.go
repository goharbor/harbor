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
	"log"
	"regexp"
	"strings"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"

	"github.com/astaxie/beego"
)

// NotificationHandler handles request on /service/notifications/, which listens to registry's events.
type NotificationHandler struct {
	beego.Controller
}

const manifestPattern = `^application/vnd.docker.distribution.manifest.v\d`

// Post handles POST request, and records audit log or refreshes cache based on event.
func (n *NotificationHandler) Post() {
	var notification models.Notification
	//log.Printf("Notification Handler triggered!\n")
	//log.Printf("request body in string: %s", string(n.Ctx.Input.CopyBody(1<<32)))
	err := json.Unmarshal(n.Ctx.Input.CopyBody(1<<32), &notification)

	if err != nil {
		beego.Error("error while decoding json: ", err)
		return
	}
	var username, action, repo, project string
	var matched bool
	for _, e := range notification.Events {
		matched, err = regexp.MatchString(manifestPattern, e.Target.MediaType)
		if err != nil {
			beego.Error("Failed to match the media type against pattern, error: ", err)
			matched = false
		}
		if matched && strings.HasPrefix(e.Request.UserAgent, "docker") {
			username = e.Actor.Name
			action = e.Action
			repo = e.Target.Repository
			if strings.Contains(repo, "/") {
				project = repo[0:strings.LastIndex(repo, "/")]
			}
			if username == "" {
				username = "anonymous"
			}
			go dao.AccessLog(username, project, repo, action)
			if action == "push" {
				go persistPushEvent(e)
			}
		}
	}
}

// persist push infomation
func persistPushEvent(e models.Event) {
	log.Printf("in gorotine\n")
	err2 := svc_utils.RefreshCatalogCache()
	if err2 != nil {
		beego.Error("Error happens when refreshing cache:", err2)
	}
	var repository models.Repository
	repository.Name = strings.Split(e.Target.Repository, "/")[1]
	repository.ProjectName = strings.Split(e.Target.Repository, "/")[0]
	repository.UserName = e.Actor.Name
	tags := getRepoTagsFromRegistry(e.Target.Repository)
	if len(tags) > 0 {
		repository.LatestTag = tags[0]
		repositoryDao, err := dao.AddOrUpdateRepository(&repository)
		if err != nil {
			beego.Error("add or update repo error: ", err)
			return
		}
		var tag models.Tag
		tag.Version = tags[0]
		tag.ProjectID = repositoryDao.ProjectID
		tag.RepositoryID = repositoryDao.Id
		dao.AddOrUpdateTag(&tag)
	}
}

func getRepoTagsFromRegistry(repoName string) []string {
	result, err := svc_utils.RegistryAPIGet(svc_utils.BuildRegistryURL(repoName, "tags", "list"), "admin")
	if err != nil {
		return []string{}
	}

	type tag struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	t := tag{}
	json.Unmarshal(result, &t)
	return t.Tags
}

// Render returns nil as it won't render any template.
func (n *NotificationHandler) Render() error {
	return nil
}
