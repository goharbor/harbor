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
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	common_quota "github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/core/api"
	quota "github.com/goharbor/harbor/src/core/api/quota"
	"github.com/goharbor/harbor/src/core/promgr"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
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

// Dump ...
func (rm *Migrator) Dump() ([]quota.ProjectInfo, error) {
	var (
		projects []quota.ProjectInfo
		wg       sync.WaitGroup
		err      error
	)

	reposInRegistry, err := api.Catalog()
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
		_, exist := repoMap[pro.Name]
		if !exist {
			repoMap[pro.Name] = []string{item}
		} else {
			repos := repoMap[pro.Name]
			repos = append(repos, item)
			repoMap[pro.Name] = repos
		}
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
				if !exist {
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
	for _, project := range projects {
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
	for _, project := range projects {
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
			log.Error(err)
			return err
		}
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
	repoClient, err := coreutils.NewRepositoryClientForUI("harbor-core", repo)
	if err != nil {
		return quota.RepoData{}, err
	}
	tags, err := repoClient.ListTag()
	if err != nil {
		return quota.RepoData{}, err
	}
	var afnbs []*models.ArtifactAndBlob
	var afs []*models.Artifact
	var blobs []*models.Blob

	for _, tag := range tags {
		_, mediaType, payload, err := repoClient.PullManifest(tag, []string{
			schema1.MediaTypeManifest,
			schema1.MediaTypeSignedManifest,
			schema2.MediaTypeManifest,
		})
		if err != nil {
			log.Error(err)
			return quota.RepoData{}, err
		}
		manifest, desc, err := registry.UnMarshal(mediaType, payload)
		if err != nil {
			log.Error(err)
			return quota.RepoData{}, err
		}
		// self
		afnb := &models.ArtifactAndBlob{
			DigestAF:   desc.Digest.String(),
			DigestBlob: desc.Digest.String(),
		}
		afnbs = append(afnbs, afnb)
		for _, layer := range manifest.References() {
			afnb := &models.ArtifactAndBlob{
				DigestAF:   desc.Digest.String(),
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
			Repo:         strings.Split(repo, "/")[1],
			Tag:          tag,
			Digest:       desc.Digest.String(),
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
