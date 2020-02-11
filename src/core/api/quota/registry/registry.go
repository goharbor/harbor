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
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	common_quota "github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api"
	quota "github.com/goharbor/harbor/src/core/api/quota"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/pkg/errors"
)

// Migrator ...
type Migrator struct {
	pm promgr.ProjectManager
}

// NewRegistryMigrator returns a new Migrator.
func NewRegistryMigrator(pm promgr.ProjectManager) quota.QuotaMigrator {
	migrator := Migrator{
		pm: pm,
	}
	return &migrator
}

// Ping ...
func (rm *Migrator) Ping() error {
	return quota.Check(api.HealthCheckerRegistry["registry"].Check)
}

// Dump ...
func (rm *Migrator) Dump() ([]quota.ProjectInfo, error) {
	var (
		projects []quota.ProjectInfo
		wg       sync.WaitGroup
		err      error
	)

	reposInRegistry, err := registry.Cli.Catalog()
	if err != nil {
		return nil, err
	}

	// repoMap : map[project_name : []repo list]
	repoMap := make(map[string][]string)
	for _, item := range reposInRegistry {
		projectName := strings.Split(item, "/")[0]
		pro, err := rm.pm.Get(projectName)
		if err != nil {
			log.Errorf("failed to get project %s: %v", projectName, err)
			continue
		}
		if pro == nil {
			continue
		}
		_, exist := repoMap[pro.Name]
		if !exist {
			repoMap[pro.Name] = []string{item}
		} else {
			repos := repoMap[pro.Name]
			repos = append(repos, item)
			repoMap[pro.Name] = repos
		}
	}
	repoMap, err = rm.appendEmptyProject(repoMap)
	if err != nil {
		log.Errorf("fail to add empty projects: %v", err)
		return nil, err
	}

	wg.Add(len(repoMap))
	errChan := make(chan error, 1)
	infoChan := make(chan interface{})
	done := make(chan bool, 1)

	go func() {
		defer func() {
			done <- true
		}()

		for {
			select {
			case result := <-infoChan:
				if result == nil {
					return
				}
				project, ok := result.(quota.ProjectInfo)
				if ok {
					projects = append(projects, project)
				}

			case e := <-errChan:
				if err == nil {
					err = errors.Wrap(e, "quota sync error on getting info of project")
				} else {
					err = errors.Wrap(e, err.Error())
				}
			}
		}
	}()

	for project, repos := range repoMap {
		go func(project string, repos []string) {
			defer wg.Done()
			info, err := infoOfProject(project, repos)
			if err != nil {
				errChan <- err
				return
			}
			infoChan <- info
		}(project, repos)
	}

	wg.Wait()
	close(infoChan)

	// wait for all of project info
	<-done

	if err != nil {
		return nil, err
	}

	return projects, nil
}

// As catalog api cannot list the empty projects in harbor, here it needs to append the empty projects into repo infor
// so that quota syncer can add 0 usage into quota usage.
func (rm *Migrator) appendEmptyProject(repoMap map[string][]string) (map[string][]string, error) {
	var withEmptyProjects map[string][]string
	all, err := dao.GetProjects(nil)
	if err != nil {
		return withEmptyProjects, err
	}
	withEmptyProjects = repoMap
	for _, pro := range all {
		_, exist := repoMap[pro.Name]
		if !exist {
			withEmptyProjects[pro.Name] = []string{}
		}
	}
	return withEmptyProjects, nil
}

// Usage ...
// registry needs to merge the shard blobs of different repositories.
func (rm *Migrator) Usage(projects []quota.ProjectInfo) ([]quota.ProjectUsage, error) {
	var pros []quota.ProjectUsage

	for _, project := range projects {
		var size, count int64
		var blobs = make(map[string]int64)

		// usage count
		for _, repo := range project.Repos {
			count = count + int64(len(repo.Afs))
			// Because that there are some shared blobs between repositories, it needs to remove the duplicate items.
			for _, blob := range repo.Blobs {
				_, exist := blobs[blob.Digest]
				// foreign blob won't be calculated
				if !exist && blob.ContentType != common.ForeignLayer {
					blobs[blob.Digest] = blob.Size
				}
			}
		}
		// size
		for _, item := range blobs {
			size = size + item
		}

		proUsage := quota.ProjectUsage{
			Project: project.Name,
			Used: common_quota.ResourceList{
				common_quota.ResourceCount:   count,
				common_quota.ResourceStorage: size,
			},
		}
		pros = append(pros, proUsage)
	}

	return pros, nil
}

// Persist ...
func (rm *Migrator) Persist(projects []quota.ProjectInfo) error {
	total := len(projects)
	for i, project := range projects {
		log.Infof("[Quota-Sync]:: start to persist artifact&blob for project: %s, progress... [%d/%d]", project.Name, i, total)
		for _, repo := range project.Repos {
			if err := persistAf(repo.Afs); err != nil {
				return err
			}
			if err := persistAfnbs(repo.Afnbs); err != nil {
				return err
			}
			if err := persistBlob(repo.Blobs); err != nil {
				return err
			}
		}
		log.Infof("[Quota-Sync]:: success to persist artifact&blob for project: %s, progress... [%d/%d]", project.Name, i, total)
	}
	if err := persistPB(projects); err != nil {
		return err
	}
	return nil
}

func persistAf(afs []*models.Artifact) error {
	if len(afs) != 0 {
		for _, af := range afs {
			_, err := dao.AddArtifact(af)
			if err != nil {
				if err == dao.ErrDupRows {
					continue
				}
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func persistAfnbs(afnbs []*models.ArtifactAndBlob) error {
	if len(afnbs) != 0 {
		for _, afnb := range afnbs {
			_, err := dao.AddArtifactNBlob(afnb)
			if err != nil {
				if err == dao.ErrDupRows {
					continue
				}
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func persistBlob(blobs []*models.Blob) error {
	if len(blobs) != 0 {
		for _, blob := range blobs {
			_, err := dao.AddBlob(blob)
			if err != nil {
				if err == dao.ErrDupRows {
					continue
				}
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func persistPB(projects []quota.ProjectInfo) error {
	total := len(projects)
	for i, project := range projects {
		log.Infof("[Quota-Sync]:: start to persist project&blob for project: %s, progress... [%d/%d]", project.Name, i, total)
		var blobs = make(map[string]int64)
		var blobsOfPro []*models.Blob
		for _, repo := range project.Repos {
			for _, blob := range repo.Blobs {
				_, exist := blobs[blob.Digest]
				if exist {
					continue
				}
				blobs[blob.Digest] = blob.Size
				blobInDB, err := dao.GetBlob(blob.Digest)
				if err != nil {
					log.Error(err)
					return err
				}
				if blobInDB != nil {
					blobsOfPro = append(blobsOfPro, blobInDB)
				}
			}
		}
		pro, err := dao.GetProjectByName(project.Name)
		if err != nil {
			log.Error(err)
			return err
		}
		_, err = dao.AddBlobsToProject(pro.ProjectID, blobsOfPro...)
		if err != nil {
			if err == dao.ErrDupRows {
				continue
			}
			log.Error(err)
			return err
		}
		log.Infof("[Quota-Sync]:: success to persist project&blob for project: %s, progress... [%d/%d]", project.Name, i, total)
	}
	return nil
}

func infoOfProject(project string, repoList []string) (quota.ProjectInfo, error) {
	var (
		repos []quota.RepoData
		wg    sync.WaitGroup
		err   error
	)
	wg.Add(len(repoList))

	errChan := make(chan error, 1)
	infoChan := make(chan interface{})
	done := make(chan bool, 1)

	pro, err := dao.GetProjectByName(project)
	if err != nil {
		log.Error(err)
		return quota.ProjectInfo{}, err
	}

	go func() {
		defer func() {
			done <- true
		}()

		for {
			select {
			case result := <-infoChan:
				if result == nil {
					return
				}
				repoData, ok := result.(quota.RepoData)
				if ok {
					repos = append(repos, repoData)
				}

			case e := <-errChan:
				if err == nil {
					err = errors.Wrap(e, "quota sync error on getting info of repo")
				} else {
					err = errors.Wrap(e, err.Error())
				}
			}
		}
	}()

	for _, repo := range repoList {
		go func(pid int64, repo string) {
			defer func() {
				wg.Done()
			}()
			info, err := infoOfRepo(pid, repo)
			if err != nil {
				errChan <- err
				return
			}
			infoChan <- info
		}(pro.ProjectID, repo)
	}

	wg.Wait()
	close(infoChan)

	<-done

	if err != nil {
		return quota.ProjectInfo{}, err
	}

	return quota.ProjectInfo{
		Name:  project,
		Repos: repos,
	}, nil
}

func infoOfRepo(pid int64, repo string) (quota.RepoData, error) {
	tags, err := registry.Cli.ListTags(repo)
	if err != nil {
		return quota.RepoData{}, err
	}
	var afnbs []*models.ArtifactAndBlob
	var afs []*models.Artifact
	var blobs []*models.Blob

	for _, tag := range tags {
		manifest, digest, err := registry.Cli.PullManifest(repo, tag)
		if err != nil {
			log.Error(err)
			// To workaround issue: https://github.com/goharbor/harbor/issues/9299, just log the error and do not raise it.
			// Let the sync process pass, but the 'Unknown manifest' will not be counted into size and count of quota usage.
			// User still can view there images with size 0 in harbor.
			continue
		}
		mediaType, payload, err := manifest.Payload()
		if err != nil {
			return quota.RepoData{}, err
		}
		// self
		afnb := &models.ArtifactAndBlob{
			DigestAF:   digest,
			DigestBlob: digest,
		}
		afnbs = append(afnbs, afnb)
		// add manifest as a blob.
		blob := &models.Blob{
			Digest:       digest,
			ContentType:  mediaType,
			Size:         int64(len(payload)),
			CreationTime: time.Now(),
		}
		blobs = append(blobs, blob)
		for _, layer := range manifest.References() {
			afnb := &models.ArtifactAndBlob{
				DigestAF:   digest,
				DigestBlob: layer.Digest.String(),
			}
			afnbs = append(afnbs, afnb)
			blob := &models.Blob{
				Digest:       layer.Digest.String(),
				ContentType:  layer.MediaType,
				Size:         layer.Size,
				CreationTime: time.Now(),
			}
			blobs = append(blobs, blob)
		}
		af := &models.Artifact{
			PID:          pid,
			Repo:         repo,
			Tag:          tag,
			Digest:       digest,
			Kind:         "Docker-Image",
			CreationTime: time.Now(),
		}
		afs = append(afs, af)
	}
	return quota.RepoData{
		Name:  repo,
		Afs:   afs,
		Afnbs: afnbs,
		Blobs: blobs,
	}, nil
}

func init() {
	quota.Register("registry", NewRegistryMigrator)
}
