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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/jobservice/logger"

	"github.com/goharbor/harbor/src/pkg/scan/api/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/notary"
	notarymodel "github.com/goharbor/harbor/src/common/utils/notary/model"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/core/config"
	notifierEvt "github.com/goharbor/harbor/src/core/notifier/event"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/model"
)

// RepositoryAPI handles request to /api/repositories /api/repositories/tags /api/repositories/manifests, the parm has to be put
// in the query string as the web framework can not parse the URL if it contains veriadic sectors.
type RepositoryAPI struct {
	BaseController
}

type repoResp struct {
	ID           int64           `json:"id"`
	Index        int             `json:"-"`
	Name         string          `json:"name"`
	ProjectID    int64           `json:"project_id"`
	Description  string          `json:"description"`
	PullCount    int64           `json:"pull_count"`
	StarCount    int64           `json:"star_count"`
	TagsCount    int64           `json:"tags_count"`
	Labels       []*models.Label `json:"labels"`
	CreationTime time.Time       `json:"creation_time"`
	UpdateTime   time.Time       `json:"update_time"`
}

type reposSorter []*repoResp

func (r reposSorter) Len() int {
	return len(r)
}

func (r reposSorter) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r reposSorter) Less(i, j int) bool {
	return r[i].Index < r[j].Index
}

type manifestResp struct {
	Manifest interface{} `json:"manifest"`
	Config   interface{} `json:"config,omitempty" `
}

// Get ...
func (ra *RepositoryAPI) Get() {
	projectID, err := ra.GetInt64("project_id")
	if err != nil || projectID <= 0 {
		ra.SendBadRequestError(fmt.Errorf("invalid project_id %s", ra.GetString("project_id")))
		return
	}

	labelID, err := ra.GetInt64("label_id", 0)
	if err != nil {
		ra.SendBadRequestError(fmt.Errorf("invalid label_id: %s", ra.GetString("label_id")))
		return
	}

	exist, err := ra.ProjectMgr.Exists(projectID)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %d",
			projectID), err)
		return
	}

	if !exist {
		ra.SendNotFoundError(fmt.Errorf("project %d not found", projectID))
		return
	}

	if !ra.RequireProjectAccess(projectID, rbac.ActionList, rbac.ResourceRepository) {
		return
	}

	query := &models.RepositoryQuery{
		ProjectIDs: []int64{projectID},
		Name:       ra.GetString("q"),
		LabelID:    labelID,
	}
	query.Page, query.Size, err = ra.GetPaginationParams()
	if err != nil {
		ra.SendBadRequestError(err)
		return
	}

	query.Sort = ra.GetString("sort")

	total, err := dao.GetTotalOfRepositories(query)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to get total of repositories of project %d: %v",
			projectID, err))
		return
	}

	repositories, err := getRepositories(query)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to get repository: %v", err))
		return
	}

	ra.SetPaginationHeader(total, query.Page, query.Size)
	ra.Data["json"] = repositories
	ra.ServeJSON()
}

func getRepositories(query *models.RepositoryQuery) ([]*repoResp, error) {
	repositories, err := dao.GetRepositories(query)
	if err != nil {
		return nil, err
	}

	return assembleReposInParallel(repositories), nil
}

func assembleReposInParallel(repositories []*models.RepoRecord) []*repoResp {
	c := make(chan *repoResp)
	for i, repository := range repositories {
		go assembleRepo(c, i, repository)
	}
	result := []*repoResp{}
	var item *repoResp
	for i := 0; i < len(repositories); i++ {
		item = <-c
		if item == nil {
			continue
		}
		result = append(result, item)
	}
	sort.Sort(reposSorter(result))

	return result
}

func assembleRepo(c chan *repoResp, index int, repository *models.RepoRecord) {
	repo := &repoResp{
		Index:        index,
		ID:           repository.RepositoryID,
		Name:         repository.Name,
		ProjectID:    repository.ProjectID,
		Description:  repository.Description,
		PullCount:    repository.PullCount,
		StarCount:    repository.StarCount,
		CreationTime: repository.CreationTime,
		UpdateTime:   repository.UpdateTime,
	}

	tags, err := getTags(repository.Name)
	if err != nil {
		log.Errorf("failed to list tags of %s: %v", repository.Name, err)
	} else {
		repo.TagsCount = int64(len(tags))
	}

	labels, err := dao.GetLabelsOfResource(common.ResourceTypeRepository, repository.RepositoryID)
	if err != nil {
		log.Errorf("failed to get labels of repository %s: %v", repository.Name, err)
	} else {
		repo.Labels = labels
	}

	c <- repo
}

// Delete ...
func (ra *RepositoryAPI) Delete() {
	// using :splat to get * part in path
	repoName := ra.GetString(":splat")

	projectName, _ := utils.ParseRepository(repoName)
	project, err := ra.ProjectMgr.Get(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to get the project %s",
			projectName), err)
		return
	}

	if project == nil {
		ra.SendNotFoundError(fmt.Errorf("project %s not found", projectName))
		return
	}

	if !ra.RequireAuthenticated() ||
		!ra.RequireProjectAccess(project.ProjectID, rbac.ActionDelete, rbac.ResourceRepository) {
		return
	}

	rc, err := coreutils.NewRepositoryClientForLocal(ra.SecurityCtx.GetUsername(), repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.SendInternalServerError(errors.New("internal error"))
		return
	}

	tags := []string{}
	tag := ra.GetString(":tag")
	if len(tag) == 0 {
		tagList, err := rc.ListTag()
		if err != nil {
			ra.ParseAndHandleError(fmt.Sprintf("error occurred while listing tags of %s", repoName), err)
			return
		}

		// TODO remove the logic if the bug of registry is fixed
		if len(tagList) == 0 {
			ra.SendNotFoundError(fmt.Errorf("no tags found for repository %s", repoName))
			return
		}

		tags = append(tags, tagList...)
	} else {
		tags = append(tags, tag)
	}

	if config.WithNotary() {
		signedTags, err := getSignatures(ra.SecurityCtx.GetUsername(), repoName)
		if err != nil {
			ra.SendInternalServerError(fmt.Errorf(
				"failed to get signatures for repository %s: %v", repoName, err))
			return
		}

		for _, t := range tags {
			digest, _, err := rc.ManifestExist(t)
			if err != nil {
				log.Errorf("Failed to Check the digest of tag: %s, error: %v", t, err.Error())
				ra.SendInternalServerError(err)
				return
			}
			log.Debugf("Tag: %s, digest: %s", t, digest)
			if _, ok := signedTags[digest]; ok {
				log.Errorf("Found signed tag, repostory: %s, tag: %s, deletion will be canceled", repoName, t)
				ra.SendPreconditionFailedError(fmt.Errorf("tag %s is signed", t))
				return
			}
		}
	}

	for _, t := range tags {
		image := fmt.Sprintf("%s:%s", repoName, t)
		if err = dao.DeleteLabelsOfResource(common.ResourceTypeImage, image); err != nil {
			ra.SendInternalServerError(fmt.Errorf("failed to delete labels of image %s: %v", image, err))
			return
		}
		if err = rc.DeleteTag(t); err != nil {
			if regErr, ok := err.(*commonhttp.Error); ok {
				if regErr.Code == http.StatusNotFound {
					continue
				}
			}
			ra.ParseAndHandleError(fmt.Sprintf("failed to delete tag %s", t), err)
			return
		}
		log.Infof("delete tag: %s:%s", repoName, t)

		go func(tag string) {
			e := &event.Event{
				Type: event.EventTypeImageDelete,
				Resource: &model.Resource{
					Type: model.ResourceTypeImage,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: repoName,
						},
						Vtags: []string{tag},
					},
					Deleted: true,
				},
			}
			if err := replication.EventHandler.Handle(e); err != nil {
				log.Errorf("failed to handle event: %v", err)
			}
		}(t)

		go func(tag string) {
			if err := dao.AddAccessLog(models.AccessLog{
				Username:  ra.SecurityCtx.GetUsername(),
				ProjectID: project.ProjectID,
				RepoName:  repoName,
				RepoTag:   tag,
				Operation: "delete",
				OpTime:    time.Now(),
			}); err != nil {
				log.Errorf("failed to add access log: %v", err)
			}
		}(t)
	}

	// build and publish image delete event
	evt := &notifierEvt.Event{}
	imgDelMetadata := &notifierEvt.ImageDelMetaData{
		Project:  project,
		Tags:     tags,
		RepoName: repoName,
		OccurAt:  time.Now(),
		Operator: ra.SecurityCtx.GetUsername(),
	}
	if err := evt.Build(imgDelMetadata); err == nil {
		if err := evt.Publish(); err != nil {
			// do not return when publishing event failed
			log.Errorf("failed to publish image delete event: %v", err)
		}
	} else {
		// do not return when building event metadata failed
		log.Errorf("failed to build image delete event metadata: %v", err)
	}

	exist, err := repositoryExist(repoName, rc)
	if err != nil {
		log.Errorf("failed to check the existence of repository %s: %v", repoName, err)
		ra.SendInternalServerError(fmt.Errorf("failed to check the existence of repository %s: %v", repoName, err))
		return
	}
	if !exist {
		repository, err := dao.GetRepositoryByName(repoName)
		if err != nil {
			ra.SendInternalServerError(fmt.Errorf("failed to get repository %s: %v", repoName, err))
			return
		}
		if repository == nil {
			log.Warningf("the repository %s not found after deleting tags", repoName)
			return
		}

		if err = dao.DeleteLabelsOfResource(common.ResourceTypeRepository,
			strconv.FormatInt(repository.RepositoryID, 10)); err != nil {
			ra.SendInternalServerError(fmt.Errorf("failed to delete labels of repository %s: %v",
				repoName, err))
			return
		}
		if err = dao.DeleteRepository(repoName); err != nil {
			log.Errorf("failed to delete repository %s: %v", repoName, err)
			ra.SendInternalServerError(fmt.Errorf("failed to delete repository %s: %v", repoName, err))
			return
		}
	}
}

// GetTag returns the tag of a repository
func (ra *RepositoryAPI) GetTag() {
	repository := ra.GetString(":splat")
	tag := ra.GetString(":tag")
	exist, _, err := ra.checkExistence(repository, tag)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to check the existence of resource, error: %v", err))
		return
	}
	if !exist {
		ra.SendNotFoundError(fmt.Errorf("resource: %s:%s not found", repository, tag))
		return
	}

	projectName, _ := utils.ParseRepository(repository)
	if !ra.RequireProjectAccess(projectName, rbac.ActionRead, rbac.ResourceRepositoryTag) {
		return
	}

	project, err := ra.ProjectMgr.Get(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to get the project %s",
			projectName), err)
		return
	}

	client, err := coreutils.NewRepositoryClientForUI(ra.SecurityCtx.GetUsername(), repository)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to initialize the client for %s: %v",
			repository, err))
		return
	}

	_, exist, err = client.ManifestExist(tag)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to check the existence of %s:%s: %v", repository, tag, err))
		return
	}
	if !exist {
		ra.SendNotFoundError(fmt.Errorf("%s not found", tag))
		return
	}

	result := assembleTagsInParallel(client, project.ProjectID, repository, []string{tag},
		ra.SecurityCtx.GetUsername())
	ra.Data["json"] = result[0]
	ra.ServeJSON()
}

// Retag tags an existing image to another tag in this repo, the source image is specified by request body.
func (ra *RepositoryAPI) Retag() {
	if !ra.SecurityCtx.IsAuthenticated() {
		ra.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	repoName := ra.GetString(":splat")
	project, repo := utils.ParseRepository(repoName)
	if !utils.ValidateRepo(repo) {
		ra.SendBadRequestError(fmt.Errorf("invalid repo '%s'", repo))
		return
	}

	request := models.RetagRequest{}
	if err := ra.DecodeJSONReq(&request); err != nil {
		ra.SendBadRequestError(err)
		return
	}
	srcImage, err := models.ParseImage(request.SrcImage)
	if err != nil {
		ra.SendBadRequestError(fmt.Errorf("invalid src image string '%s', should in format '<project>/<repo>:<tag>'", request.SrcImage))
		return
	}

	if !utils.ValidateTag(request.Tag) {
		ra.SendBadRequestError(fmt.Errorf("invalid tag '%s'", request.Tag))
		return
	}

	// Check whether source image exists
	exist, _, err := ra.checkExistence(fmt.Sprintf("%s/%s", srcImage.Project, srcImage.Repo), srcImage.Tag)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("check existence of %s error: %v", request.SrcImage, err))
		return
	}
	if !exist {
		ra.SendNotFoundError(fmt.Errorf("image %s not exist", request.SrcImage))
		return
	}

	// Check whether target project exists
	exist, err = ra.ProjectMgr.Exists(project)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s", project), err)
		return
	}
	if !exist {
		ra.SendNotFoundError(fmt.Errorf("project %s not found", project))
		return
	}

	// If override not allowed, check whether target tag already exists
	if !request.Override {
		exist, _, err := ra.checkExistence(repoName, request.Tag)
		if err != nil {
			ra.SendInternalServerError(fmt.Errorf("check existence of %s:%s error: %v", repoName, request.Tag, err))
			return
		}
		if exist {
			ra.SendConflictError(fmt.Errorf("tag '%s' already existed for '%s'", request.Tag, repoName))
			return
		}
	}

	// Check whether user has read permission to source project
	if hasPermission, _ := ra.HasProjectPermission(srcImage.Project, rbac.ActionPull, rbac.ResourceRepository); !hasPermission {
		log.Errorf("user has no read permission to project '%s'", srcImage.Project)
		ra.SendForbiddenError(fmt.Errorf("%s has no read permission to project %s", ra.SecurityCtx.GetUsername(), srcImage.Project))
		return
	}

	// Check whether user has write permission to target project
	if hasPermission, _ := ra.HasProjectPermission(project, rbac.ActionPush, rbac.ResourceRepository); !hasPermission {
		log.Errorf("user has no write permission to project '%s'", project)
		ra.SendForbiddenError(fmt.Errorf("%s has no write permission to project %s", ra.SecurityCtx.GetUsername(), project))
		return
	}

	// Retag the image
	if err = coreutils.Retag(srcImage, &models.Image{
		Project: project,
		Repo:    repo,
		Tag:     request.Tag,
	}); err != nil {
		if e, ok := err.(*commonhttp.Error); ok {
			ra.RenderFormattedError(e.Code, e.Message)
			return
		}
		ra.SendInternalServerError(fmt.Errorf("%v", err))
	}
}

// GetTags returns tags of a repository
func (ra *RepositoryAPI) GetTags() {
	repoName := ra.GetString(":splat")
	labelID, err := ra.GetInt64("label_id", 0)
	if err != nil {
		ra.SendBadRequestError(fmt.Errorf("invalid label_id: %s", ra.GetString("label_id")))
		return
	}

	projectName, _ := utils.ParseRepository(repoName)
	project, err := ra.ProjectMgr.Get(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to get the project %s",
			projectName), err)
		return
	}

	if project == nil {
		ra.SendNotFoundError(fmt.Errorf("project %s not found", projectName))
		return
	}

	if !ra.RequireProjectAccess(projectName, rbac.ActionList, rbac.ResourceRepositoryTag) {
		return
	}

	client, err := coreutils.NewRepositoryClientForUI(ra.SecurityCtx.GetUsername(), repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.SendInternalServerError(errors.New("internal error"))
		return
	}

	tags, err := client.ListTag()
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to get tag of %s: %v", repoName, err))
		return
	}

	// filter tags by label ID
	if labelID > 0 {
		rls, err := dao.ListResourceLabels(&models.ResourceLabelQuery{
			LabelID:      labelID,
			ResourceType: common.ResourceTypeImage,
		})
		if err != nil {
			ra.SendInternalServerError(fmt.Errorf("failed to list resource labels: %v", err))
			return
		}
		labeledTags := map[string]struct{}{}
		for _, rl := range rls {
			strs := strings.SplitN(rl.ResourceName, ":", 2)
			// the "rls" may contain images which don't belong to the repository
			if strs[0] != repoName {
				continue
			}
			labeledTags[strs[1]] = struct{}{}
		}
		ts := []string{}
		for _, tag := range tags {
			if _, ok := labeledTags[tag]; ok {
				ts = append(ts, tag)
			}
		}
		tags = ts
	}

	detail, err := ra.GetBool("detail", true)
	if !detail && err == nil {
		ra.Data["json"] = simpleTags(tags)
		ra.ServeJSON()
		return
	}

	ra.Data["json"] = assembleTagsInParallel(
		client,
		project.ProjectID,
		repoName,
		tags,
		ra.SecurityCtx.GetUsername(),
	)
	ra.ServeJSON()
}

func simpleTags(tags []string) []*models.TagResp {
	var tagsResp []*models.TagResp
	for _, tag := range tags {
		tagsResp = append(tagsResp, &models.TagResp{
			TagDetail: models.TagDetail{
				Name: tag,
			},
		})
	}

	return tagsResp
}

// get config, signature and scan overview and assemble them into one
// struct for each tag in tags
func assembleTagsInParallel(client *registry.Repository, projectID int64, repository string,
	tags []string, username string) []*models.TagResp {
	var err error
	signatures := map[string][]notarymodel.Target{}
	if config.WithNotary() {
		signatures, err = getSignatures(username, repository)
		if err != nil {
			signatures = map[string][]notarymodel.Target{}
			log.Errorf("failed to get signatures of %s: %v", repository, err)
		}
	}

	c := make(chan *models.TagResp)
	for _, tag := range tags {
		go assembleTag(
			c,
			client,
			projectID,
			repository,
			tag,
			config.WithNotary(),
			signatures,
		)
	}
	result := []*models.TagResp{}
	var item *models.TagResp
	for i := 0; i < len(tags); i++ {
		item = <-c
		if item == nil {
			continue
		}
		result = append(result, item)
	}
	return result
}

func assembleTag(c chan *models.TagResp, client *registry.Repository, projectID int64,
	repository, tag string, notaryEnabled bool,
	signatures map[string][]notarymodel.Target) {
	item := &models.TagResp{}
	// labels
	image := fmt.Sprintf("%s:%s", repository, tag)
	labels, err := dao.GetLabelsOfResource(common.ResourceTypeImage, image)
	if err != nil {
		log.Errorf("failed to get labels of image %s: %v", image, err)
	} else {
		item.Labels = labels
	}

	// the detail information of tag
	tagDetail, err := getTagDetail(client, tag)
	if err != nil {
		log.Errorf("failed to get v2 manifest of %s:%s: %v", repository, tag, err)
	}
	if tagDetail != nil {
		item.TagDetail = *tagDetail
	}

	// scan overview
	so := getSummary(projectID, repository, item.Digest)
	if len(so) > 0 {
		item.ScanOverview = so
	}

	// signature, compare both digest and tag
	if notaryEnabled && signatures != nil {
		if sigs, ok := signatures[item.Digest]; ok {
			for _, sig := range sigs {
				if item.Name == sig.Tag {
					item.Signature = &sig
				}
			}
		}
	}

	// pull/push time
	artifact, err := dao.GetArtifact(repository, tag)
	if err != nil {
		log.Errorf("failed to get artifact %s:%s: %v", repository, tag, err)
	} else {
		if artifact == nil {
			log.Warningf("artifact %s:%s not found", repository, tag)
		} else {
			item.PullTime = artifact.PullTime
			item.PushTime = artifact.PushTime
		}
	}

	c <- item
}

// getTagDetail returns the detail information for v2 manifest image
// The information contains architecture, os, author, size, etc.
func getTagDetail(client *registry.Repository, tag string) (*models.TagDetail, error) {
	detail := &models.TagDetail{
		Name: tag,
	}

	digest, mediaType, payload, err := client.PullManifest(tag, []string{schema2.MediaTypeManifest})
	if err != nil {
		return detail, err
	}
	detail.Digest = digest

	if strings.Contains(mediaType, "application/json") {
		mediaType = schema1.MediaTypeManifest
	}
	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		return detail, err
	}

	// size of manifest + size of layers
	detail.Size = int64(len(payload))
	for _, ref := range manifest.References() {
		detail.Size += ref.Size
	}

	// if the media type of the manifest isn't v2, doesn't parse image config
	// and return directly
	// this impacts that some detail information(os, arch, ...) of old images
	// cannot be got
	if mediaType != schema2.MediaTypeManifest {
		log.Debugf("the media type of the manifest is %s, not v2, skip", mediaType)
		return detail, nil
	}
	v2Manifest, ok := manifest.(*schema2.DeserializedManifest)
	if !ok {
		log.Debug("the manifest cannot be convert to DeserializedManifest, skip")
		return detail, nil
	}

	_, reader, err := client.PullBlob(v2Manifest.Target().Digest.String())
	if err != nil {
		return detail, err
	}

	configData, err := ioutil.ReadAll(reader)
	if err != nil {
		return detail, err
	}

	if err = json.Unmarshal(configData, detail); err != nil {
		return detail, err
	}

	populateAuthor(detail)

	return detail, nil
}

func populateAuthor(detail *models.TagDetail) {
	// has author info already
	if len(detail.Author) > 0 {
		return
	}

	// try to set author with the value of label "maintainer"
	if detail.Config != nil {
		for k, v := range detail.Config.Labels {
			if strings.ToLower(k) == "maintainer" {
				detail.Author = v
				return
			}
		}
	}
}

// GetManifests returns the manifest of a tag
func (ra *RepositoryAPI) GetManifests() {
	repoName := ra.GetString(":splat")
	tag := ra.GetString(":tag")

	version := ra.GetString("version")
	if len(version) == 0 {
		version = "v2"
	}

	if version != "v1" && version != "v2" {
		ra.SendBadRequestError(errors.New("version should be v1 or v2"))
		return
	}

	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exists(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			projectName), err)
		return
	}

	if !exist {
		ra.SendNotFoundError(fmt.Errorf("project %s not found", projectName))
		return
	}

	if !ra.RequireProjectAccess(projectName, rbac.ActionRead, rbac.ResourceRepositoryTagManifest) {
		return
	}

	rc, err := coreutils.NewRepositoryClientForUI(ra.SecurityCtx.GetUsername(), repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.SendInternalServerError(errors.New("internal error"))
		return
	}

	manifest, err := getManifest(rc, tag, version)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("error occurred while getting manifest of %s:%s", repoName, tag), err)
		return
	}

	ra.Data["json"] = manifest
	ra.ServeJSON()
}

func getManifest(client *registry.Repository,
	tag, version string) (*manifestResp, error) {
	result := &manifestResp{}

	mediaTypes := []string{}
	switch version {
	case "v1":
		mediaTypes = append(mediaTypes, schema1.MediaTypeManifest)
	case "v2":
		mediaTypes = append(mediaTypes, schema2.MediaTypeManifest)
	}

	_, mediaType, payload, err := client.PullManifest(tag, mediaTypes)
	if err != nil {
		return nil, err
	}

	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		return nil, err
	}

	result.Manifest = manifest

	deserializedmanifest, ok := manifest.(*schema2.DeserializedManifest)
	if ok {
		_, data, err := client.PullBlob(deserializedmanifest.Target().Digest.String())
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(data)
		if err != nil {
			return nil, err
		}

		result.Config = string(b)
	}

	return result, nil
}

// GetTopRepos returns the most populor repositories
func (ra *RepositoryAPI) GetTopRepos() {
	count, err := ra.GetInt("count", 10)
	if err != nil || count <= 0 {
		ra.SendBadRequestError(fmt.Errorf("invalid count: %s", ra.GetString("count")))
		return
	}

	projectIDs := []int64{}
	projects, err := ra.ProjectMgr.GetPublic()
	if err != nil {
		ra.ParseAndHandleError("failed to get public projects", err)
		return
	}
	if ra.SecurityCtx.IsAuthenticated() {
		list, err := ra.SecurityCtx.GetMyProjects()
		if err != nil {
			ra.SendInternalServerError(fmt.Errorf("failed to get projects which the user %s is a member of: %v",
				ra.SecurityCtx.GetUsername(), err))
			return
		}
		projects = append(projects, list...)
	}

	for _, project := range projects {
		projectIDs = append(projectIDs, project.ProjectID)
	}

	repos, err := dao.GetTopRepos(projectIDs, count)
	if err != nil {
		log.Errorf("failed to get top repos: %v", err)
		ra.SendInternalServerError(errors.New("internal server error"))
		return
	}

	ra.Data["json"] = assembleReposInParallel(repos)
	ra.ServeJSON()
}

// Put updates description info for the repository
func (ra *RepositoryAPI) Put() {
	name := ra.GetString(":splat")
	repository, err := dao.GetRepositoryByName(name)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to get repository %s: %v", name, err))
		return
	}

	if repository == nil {
		ra.SendNotFoundError(fmt.Errorf("repository %s not found", name))
		return
	}

	if !ra.SecurityCtx.IsAuthenticated() {
		ra.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	projectName, _ := utils.ParseRepository(name)
	if !ra.RequireProjectAccess(projectName, rbac.ActionUpdate, rbac.ResourceRepository) {
		return
	}

	desc := struct {
		Description string `json:"description"`
	}{}
	if err := ra.DecodeJSONReq(&desc); err != nil {
		ra.SendBadRequestError(err)
		return
	}

	repository.Description = desc.Description
	if err = dao.UpdateRepository(*repository); err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to update repository %s: %v", name, err))
		return
	}
}

// GetSignatures returns signatures of a repository
func (ra *RepositoryAPI) GetSignatures() {
	repoName := ra.GetString(":splat")

	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exists(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			projectName), err)
		return
	}

	if !exist {
		ra.SendNotFoundError(fmt.Errorf("project %s not found", projectName))
		return
	}

	if !ra.RequireProjectAccess(projectName, rbac.ActionRead, rbac.ResourceRepository) {
		return
	}

	targets, err := notary.GetInternalTargets(config.InternalNotaryEndpoint(),
		ra.SecurityCtx.GetUsername(), repoName)
	if err != nil {
		log.Errorf("Error while fetching signature from notary: %v", err)
		ra.SendInternalServerError(errors.New("internal error"))
		return
	}
	ra.Data["json"] = targets
	ra.ServeJSON()
}

func getSignatures(username, repository string) (map[string][]notarymodel.Target, error) {
	targets, err := notary.GetInternalTargets(config.InternalNotaryEndpoint(),
		username, repository)
	if err != nil {
		return nil, err
	}

	signatures := map[string][]notarymodel.Target{}
	for _, tgt := range targets {
		digest, err := notary.DigestFromTarget(tgt)
		if err != nil {
			return nil, err
		}
		signatures[digest] = append(signatures[digest], tgt)
	}

	return signatures, nil
}

func (ra *RepositoryAPI) checkExistence(repository, tag string) (bool, string, error) {
	project, _ := utils.ParseRepository(repository)
	exist, err := ra.ProjectMgr.Exists(project)
	if err != nil {
		return false, "", err
	}
	if !exist {
		log.Errorf("project %s not found", project)
		return false, "", nil
	}
	client, err := coreutils.NewRepositoryClientForUI(ra.SecurityCtx.GetUsername(), repository)
	if err != nil {
		return false, "", fmt.Errorf("failed to initialize the client for %s: %v", repository, err)
	}
	digest, exist, err := client.ManifestExist(tag)
	if err != nil {
		return false, "", fmt.Errorf("failed to check the existence of %s:%s: %v", repository, tag, err)
	}
	if !exist {
		log.Errorf("%s not found", tag)
		return false, "", nil
	}
	return true, digest, nil
}

func getSummary(pid int64, repository string, digest string) map[string]interface{} {
	// At present, only get harbor native report as default behavior.
	artifact := &v1.Artifact{
		NamespaceID: pid,
		Repository:  repository,
		Digest:      digest,
		MimeType:    v1.MimeTypeDockerArtifact,
	}

	sum, err := scan.DefaultController.GetSummary(artifact, []string{v1.MimeTypeNativeReport})
	if err != nil {
		logger.Errorf("Failed to get scan report summary with error: %s", err)
	}

	return sum
}
