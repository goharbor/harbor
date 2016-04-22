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
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry"
	"github.com/vmware/harbor/utils/registry/errors"

	"github.com/astaxie/beego"
)

// NotificationHandler handles request on /service/notifications/, which listens to registry's events.
type NotificationHandler struct {
	beego.Controller
	registry *registry.Registry
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
	var username, action, repo, project, repoTag, tagURL, digest string
	var matched bool
	var client *http.Client
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
			tagURL = e.Target.URL
			digest = e.Target.Digest

			client = registry.NewClientUsernameAuthHandlerEmbeded(username)
			log.Debug("initializing username auth handler: %s", username)
			endpoint := os.Getenv("REGISTRY_URL")
			r, err1 := registry.New(endpoint, client)
			if err1 != nil {
				log.Fatalf("error occurred while initializing auth handler for repository API: %v", err1)

			}
			n.registry = r

			_, _, payload, err2 := n.registry.PullManifest(repo, digest, registry.ManifestVersion1)

			if err2 != nil {
				log.Errorf("Failed to get manifests for repo, repo name: %s, tag: %s, error: %v", repo, tagURL, err2)
				return
			}

			maniDig := models.ManifestDigest{}
			err = json.Unmarshal(payload, &maniDig)
			if err != nil {
				log.Errorf("Failed to decode json from response for manifests, repo name: %s, tag: %s, error: %v", repo, tagURL, err)
				return
			}

			var digestLayers []string
			var tagLayers []string
			for _, diglayer := range maniDig.Layers {
				digestLayers = append(digestLayers, diglayer.Digest)
			}

			tags, err := n.registry.ListTag(repo)
			if err != nil {
				e, ok := errors.ParseError(err)
				if ok {
					log.Info(e)
				} else {
					log.Error(err)
				}
				return
			}

			log.Infof("tags : %v ", tags)

			for _, tag := range tags {
				_, _, payload, err := n.registry.PullManifest(repo, tag, registry.ManifestVersion1)
				if err != nil {
					e, ok := errors.ParseError(err)
					if ok {
						log.Info(e)
					} else {
						log.Error(err)
					}
					continue
				}
				taginfo := models.Manifest{}
				err = json.Unmarshal(payload, &taginfo)
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
					repoTag = tag
					break
				}

			}

			if strings.Contains(repo, "/") {
				project = repo[0:strings.LastIndex(repo, "/")]
			}
			if username == "" {
				username = "anonymous"
			}
			log.Debugf("repo tag is : %v ", repoTag)
			go dao.AccessLog(username, project, repo, repoTag, action)
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
