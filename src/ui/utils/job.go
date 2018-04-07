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

// Package utils contains methods to support security, cache, and webhook functions.
package utils

import (
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/job"
	jobmodels "github.com/vmware/harbor/src/common/job/models"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/ui/config"

	"encoding/json"
	"fmt"
	"sync"
)

var (
	cl               sync.Mutex
	jobServiceClient job.Client
)

// ScanAllImages scans all images of Harbor by submiting jobs to jobservice, the whole process will move on if failed to submit any job of a single image.
func ScanAllImages() error {
	repos, err := dao.GetRepositories()
	if err != nil {
		log.Errorf("Failed to list all repositories, error: %v", err)
		return err
	}
	log.Infof("Scanning all images on Harbor.")

	go scanRepos(repos)
	return nil
}

// ScanImagesByProjectID scans all images under a projet, the whole process will move on if failed to submit any job of a single image.
func ScanImagesByProjectID(id int64) error {
	repos, err := dao.GetRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{id},
	})
	if err != nil {
		log.Errorf("Failed list repositories in project %d, error: %v", id, err)
		return err
	}
	log.Infof("Scanning all images in project: %d ", id)
	go scanRepos(repos)
	return nil
}

func scanRepos(repos []*models.RepoRecord) {
	var repoClient *registry.Repository
	var err error
	var tags []string
	for _, r := range repos {
		repoClient, err = NewRepositoryClientForUI("harbor-ui", r.Name)
		if err != nil {
			log.Errorf("Failed to initialize client for repository: %s, error: %v, skip scanning", r.Name, err)
			continue
		}
		tags, err = repoClient.ListTag()
		if err != nil {
			log.Errorf("Failed to get tags for repository: %s, error: %v, skip scanning.", r.Name, err)
			continue
		}
		for _, t := range tags {
			if err = TriggerImageScan(r.Name, t); err != nil {
				log.Errorf("Failed to scan image with repository: %s, tag: %s, error: %v.", r.Name, t, err)
			} else {
				log.Debugf("Triggered scan for image with repository: %s, tag: %s", r.Name, t)
			}
		}
	}
}

// GetJobServiceClient returns the job service client instance.
func GetJobServiceClient() job.Client {
	cl.Lock()
	defer cl.Unlock()
	if jobServiceClient == nil {
		jobServiceClient = job.NewDefaultClient(config.InternalJobServiceURL(), config.UISecret())
	}
	return jobServiceClient
}

// TriggerImageScan triggers an image scan job on jobservice.
func TriggerImageScan(repository string, tag string) error {
	repoClient, err := NewRepositoryClientForUI("harbor-ui", repository)
	if err != nil {
		return err
	}
	digest, _, err := repoClient.ManifestExist(tag)
	if err != nil {
		log.Errorf("Failed to get Manifest for %s:%s", repository, tag)
		return err
	}
	return triggerImageScan(repository, tag, digest, GetJobServiceClient())
}

func triggerImageScan(repository, tag, digest string, client job.Client) error {
	id, err := dao.AddScanJob(models.ScanJob{
		Repository: repository,
		Digest:     digest,
		Tag:        tag,
		Status:     models.JobPending,
	})
	if err != nil {
		return err
	}
	err = dao.SetScanJobForImg(digest, id)
	if err != nil {
		return err
	}
	data, err := buildScanJobData(id, repository, tag, digest)
	if err != nil {
		return err
	}
	uuid, err := client.SubmitJob(data)
	if err != nil {
		return err
	}
	err = dao.SetScanJobUUID(id, uuid)
	if err != nil {
		log.Warningf("Failed to set UUID for scan job, ID: %d, repository: %s, tag: %s")
	}
	return nil
}

func buildScanJobData(jobID int64, repository, tag, digest string) (*jobmodels.JobData, error) {
	parms := job.ScanJobParms{
		JobID:      jobID,
		Repository: repository,
		Digest:     digest,
		Tag:        tag,
	}
	parmsMap := make(map[string]interface{})
	b, err := json.Marshal(parms)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &parmsMap)
	if err != nil {
		return nil, err
	}
	meta := jobmodels.JobMetadata{
		JobKind:  job.JobKindGeneric,
		IsUnique: false,
	}

	data := &jobmodels.JobData{
		Name:       job.ImageScanJob,
		Parameters: jobmodels.Parameters(parmsMap),
		Metadata:   &meta,
		StatusHook: fmt.Sprintf("%s/service/notifications/jobs/scan/%d", config.InternalUIURL(), jobID),
	}

	return data, nil
}
