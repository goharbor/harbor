// Copyright 2018 Project Harbor Authors
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

package registry

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/core/service/notifications"

	"github.com/goharbor/harbor/src/api/scan"
	"github.com/goharbor/harbor/src/api/scanner"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	notifierEvt "github.com/goharbor/harbor/src/pkg/notifier/event"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/adapter"
	repevent "github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/pkg/errors"
)

// NotificationHandler handles request on /service/notifications/, which listens to registry's events.
type NotificationHandler struct {
	notifications.BaseHandler
}

const manifestPattern = `^application/vnd.docker.distribution.manifest.v\d\+(json|prettyjws)`

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

		pro, err := config.GlobalProjectMgr.Get(project)
		if err != nil {
			log.Errorf("failed to get project by name %s: %v", project, err)
			return
		}
		if pro == nil {
			log.Warningf("project %s not found", project)
			continue
		}

		go func() {
			if err := dao.AddAccessLog(models.AccessLog{
				Username:  user,
				ProjectID: pro.ProjectID,
				RepoName:  repository,
				RepoTag:   tag,
				Operation: action,
				OpTime:    time.Now(),
			}); err != nil {
				log.Errorf("failed to add access log: %v", err)
			}
		}()

		if action == "push" {
			// discard the notification without tag.
			if tag != "" {
				go func() {
					exist := dao.RepositoryExists(repository)
					if exist {
						return
					}
					log.Debugf("Add repository %s into DB.", repository)
					repoRecord := models.RepoRecord{
						Name:      repository,
						ProjectID: pro.ProjectID,
					}
					if err := dao.AddRepository(repoRecord); err != nil {
						log.Errorf("Error happens when adding repository: %v", err)
					}
				}()
			}

			if !coreutils.WaitForManifestReady(repository, tag, 6) {
				log.Errorf("Manifest for image %s:%s is not ready, skip the follow up actions.", repository, tag)
				return
			}

			// build and publish image push event
			evt := &notifierEvt.Event{}
			imgPushMetadata := &notifierEvt.ImagePushMetaData{
				Project:  pro,
				Tag:      tag,
				Digest:   event.Target.Digest,
				RepoName: event.Target.Repository,
				OccurAt:  time.Now(),
				Operator: event.Actor.Name,
			}
			if err := evt.Build(imgPushMetadata); err == nil {
				if err := evt.Publish(); err != nil {
					// do not return when publishing event failed
					log.Errorf("failed to publish image push event: %v", err)
				}
			} else {
				// do not return when building event metadata failed
				log.Errorf("failed to build image push event metadata: %v", err)
			}

			go func() {
				e := &repevent.Event{
					Type: repevent.EventTypeImagePush,
					Resource: &model.Resource{
						Type: model.ResourceTypeImage,
						Metadata: &model.ResourceMetadata{
							Repository: &model.Repository{
								Name: repository,
								Metadata: map[string]interface{}{
									"public": strconv.FormatBool(pro.IsPublic()),
								},
							},
							Vtags: []string{tag},
						},
					},
				}
				if err := replication.EventHandler.Handle(e); err != nil {
					log.Errorf("failed to handle event: %v", err)
				}
			}()

			if autoScanEnabled(pro) {
				artifact := &v1.Artifact{
					NamespaceID: pro.ProjectID,
					Repository:  repository,
					Tag:         tag,
					MimeType:    v1.MimeTypeDockerArtifact,
					Digest:      event.Target.Digest,
				}

				if err := scan.DefaultController.Scan(artifact); err != nil {
					log.Error(errors.Wrap(err, "registry notification: trigger scan when pushing automatically"))
				}
			}
		}
		if action == "pull" {
			// build and publish image pull event
			evt := &notifierEvt.Event{}
			imgPullMetadata := &notifierEvt.ImagePullMetaData{
				Project:  pro,
				Tag:      tag,
				Digest:   event.Target.Digest,
				RepoName: event.Target.Repository,
				OccurAt:  time.Now(),
				Operator: event.Actor.Name,
			}
			if err := evt.Build(imgPullMetadata); err == nil {
				if err := evt.Publish(); err != nil {
					// do not return when publishing event failed
					log.Errorf("failed to publish image pull event: %v", err)
				}
			} else {
				// do not return when building event metadata failed
				log.Errorf("failed to build image push event metadata: %v", err)
			}

			go func() {
				log.Debugf("Increase the repository %s pull count.", repository)
				if err := dao.IncreasePullCount(repository); err != nil {
					log.Errorf("Error happens when increasing pull count: %v", repository)
				}
			}()

			// update the artifact pull time, and ignore the events without tag.
			if tag != "" {
				go func() {
					artifactQuery := &models.ArtifactQuery{
						PID:  pro.ProjectID,
						Repo: repository,
					}

					// handle pull by tag or digest
					pullByDigest := utils.IsDigest(tag)
					if pullByDigest {
						artifactQuery.Digest = tag
					} else {
						artifactQuery.Tag = tag
					}

					afs, err := dao.ListArtifacts(artifactQuery)
					if err != nil {
						log.Errorf("Error occurred when to get artifact %v", err)
						return
					}
					if len(afs) > 0 {
						log.Warningf("get multiple artifact records when to update pull time with query :%d-%s-%s, "+
							"all of them will be updated.", artifactQuery.PID, artifactQuery.Repo, artifactQuery.Tag)
					}

					// ToDo: figure out how to do batch update in Pg as beego orm doesn't support update multiple like insert does.
					for _, af := range afs {
						log.Debugf("Update the artifact: %s pull time.", af.Repo)
						af.PullTime = time.Now()
						if err := dao.UpdateArtifactPullTime(af); err != nil {
							log.Errorf("Error happens when updating the pull time of artifact: %d-%s, with err: %v",
								artifactQuery.PID, artifactQuery.Repo, err)
						}
					}
				}()
			}

		}
	}
}

func filterEvents(notification *models.Notification) ([]*models.Event, error) {
	events := make([]*models.Event, 0)

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

		if checkEvent(&event) {
			events = append(events, &event)
			log.Debugf("add event to collection: %s", event.ID)
			continue
		}
	}

	return events, nil
}

func checkEvent(event *models.Event) bool {
	// push action
	if event.Action == "push" {
		return true
	}
	// if it is pull action, check the user-agent
	userAgent := strings.ToLower(strings.TrimSpace(event.Request.UserAgent))
	if userAgent == "harbor-registry-client" || userAgent == strings.ToLower(adapter.UserAgentReplication) {
		return false
	}
	return true
}

func autoScanEnabled(project *models.Project) bool {
	r, err := scanner.DefaultController.GetRegistrationByProject(project.ProjectID)
	if err != nil {
		log.Error(errors.Wrap(err, "check auto scan enable"))
		return false
	}

	// In case
	if r == nil {
		log.Errorf("no scanner is available for project: %s", project.Name)
		return false
	}

	return !r.Disabled && project.AutoScan()
}

// Render returns nil as it won't render any template.
func (n *NotificationHandler) Render() error {
	return nil
}
