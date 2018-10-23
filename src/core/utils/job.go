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

// Package utils contains methods to support security, cache, and webhook functions.
package utils

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job"
	jobmodels "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"

	"encoding/json"
	"fmt"
	"sync"
)

var (
	cl               sync.Mutex
	jobServiceClient job.Client
)

// ScanAllImages scans all images of Harbor by submiting a scan all job to jobservice, and the job handler will call API
// on the "core" service
func ScanAllImages() error {
	_, err := scanAll("")
	return err
}

// ScheduleScanAllImages will schedule a scan all job based on the cron string, add append a record in admin job table.
func ScheduleScanAllImages(cron string) error {
	_, err := scanAll(cron)
	return err
}

func scanAll(cron string, c ...job.Client) (string, error) {
	var client job.Client
	if c == nil || len(c) == 0 {
		client = GetJobServiceClient()
	} else {
		client = c[0]
	}
	kind := job.JobKindGeneric
	if len(cron) > 0 {
		kind = job.JobKindPeriodic
	}
	meta := &jobmodels.JobMetadata{
		JobKind:  kind,
		IsUnique: true,
		Cron:     cron,
	}
	id, err := dao.AddAdminJob(&models.AdminJob{
		Name: job.ImageScanAllJob,
		Kind: kind,
	})
	if err != nil {
		return "", err
	}
	data := &jobmodels.JobData{
		Name:       job.ImageScanAllJob,
		Metadata:   meta,
		StatusHook: fmt.Sprintf("%s/service/notifications/jobs/adminjob/%d", config.InternalCoreURL(), id),
	}
	log.Infof("scan_all job scheduled/triggered, cron string: '%s'", cron)
	return client.SubmitJob(data)
}

// GetJobServiceClient returns the job service client instance.
func GetJobServiceClient() job.Client {
	cl.Lock()
	defer cl.Unlock()
	if jobServiceClient == nil {
		jobServiceClient = job.NewDefaultClient(config.InternalJobServiceURL(), config.CoreSecret())
	}
	return jobServiceClient
}

// TriggerImageScan triggers an image scan job on jobservice.
func TriggerImageScan(repository string, tag string) error {
	repoClient, err := NewRepositoryClientForUI("harbor-core", repository)
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
		log.Warningf("Failed to set UUID for scan job, ID: %d, repository: %s, tag: %s", id, uuid, repository, tag)
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
		StatusHook: fmt.Sprintf("%s/service/notifications/jobs/scan/%d", config.InternalCoreURL(), jobID),
	}

	return data, nil
}
