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
	"sort"
	"strings"

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

type taglist struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
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
	var username, action, repo, project, repo_tag, tag_url string
	var matched bool
	for _, e := range notification.Events {
		matched, err = regexp.MatchString(manifestPattern, e.Target.MediaType)
		if err != nil {
			log.Errorf("Failed to match the media type against pattern, error: %v", err)
			matched = false
		}
		if matched && strings.HasPrefix(e.Request.UserAgent, "docker") {
			username = e.Actor.Name
			action = e.Action
			repo = e.Target.Repository
			tag_url = e.Target.URL
			result, err1 := svc_utils.RegistryAPIGet(tag_url, username)

			if err1 != nil {
				log.Errorf("Failed to get manifests for repo, repo name: %s, tag: %s, error: %v", repo, tag_url, err1)
				return
			}

			maniDig := models.ManifestDigest{}
			err = json.Unmarshal(result, &maniDig)
			if err != nil {
				log.Errorf("Failed to decode json from response for manifests, repo name: %s, tag: %s, error: %v", repo, tag_url, err)
				return
			}

			var digestLayers []string
			var tagLayers []string
			for _, diglayer := range maniDig.Layers {
				digestLayers = append(digestLayers, diglayer.Digest)
			}

			result, err = svc_utils.RegistryAPIGet(svc_utils.BuildRegistryURL(repo, "tags", "list"), username)
			if err != nil {
				log.Errorf("Failed to get repo tags, repo name: %s, error: %v", repo, err)
			} else {
				t := taglist{}
				json.Unmarshal(result, &t)
				for _, tag := range t.Tags {
					result, err = svc_utils.RegistryAPIGet(svc_utils.BuildRegistryURL(repo, "manifests", tag), username)
					if err != nil {
						log.Errorf("Failed to get repo tags, repo name: %s, error: %v", repo, err)
						continue
					}
					taginfo := models.Manifest{}
					err = json.Unmarshal(result, &taginfo)
					if err != nil {
						log.Errorf("Failed to decode json from response for manifests, repo name: %s, tag: %s, error: %v", repo, tag, err)
						continue
					}
					for _, fslayer := range taginfo.FsLayers {
						tagLayers = append(tagLayers, fslayer.BlobSum)
					}

					sort.Strings(digestLayers)
					sort.Strings(tagLayers)
					eq := compStringArray(digestLayers, tagLayers)
					if eq {
						repo_tag = tag
						break
					}
				}
			}

			if strings.Contains(repo, "/") {
				project = repo[0:strings.LastIndex(repo, "/")]
			}
			if username == "" {
				username = "anonymous"
			}
			log.Debugf("repo tag is : %v ", repo_tag)
			go dao.AccessLog(username, project, repo, repo_tag, action)
			if action == "push" {
				go func() {
					err2 := svc_utils.RefreshCatalogCache()
					if err2 != nil {
						log.Errorf("Error happens when refreshing cache: %v", err2)
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

func compStringArray(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
