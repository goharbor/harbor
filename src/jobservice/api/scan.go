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

package api

import (
	"net/http"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/jobservice/config"
	"github.com/vmware/harbor/src/jobservice/job"
	"github.com/vmware/harbor/src/jobservice/utils"
)

// ImageScanJob handles /api/imageScanJobs /api/imageScanJobs/:id/log
type ImageScanJob struct {
	jobBaseAPI
}

// Prepare ...
func (isj *ImageScanJob) Prepare() {
	isj.authenticate()
}

// Post creates a scanner job and hand it to statemachine.
func (isj *ImageScanJob) Post() {
	var data models.ImageScanReq
	isj.DecodeJSONReq(&data)
	log.Debugf("data: %+v", data)
	regURL, err := config.LocalRegURL()
	if err != nil {
		log.Errorf("Failed to read regURL, error: %v", err)
		isj.RenderError(http.StatusInternalServerError, "Failed to read registry URL from config")
		return
	}
	c := &http.Cookie{Name: models.UISecretCookie, Value: config.JobserviceSecret()}
	repoClient, err := utils.NewRepositoryClient(regURL, false, auth.NewCookieCredential(c),
		config.InternalTokenServiceEndpoint(), data.Repo, "pull", "push", "*")
	if err != nil {
		log.Errorf("An error occurred while creating repository client: %v", err)
		isj.RenderError(http.StatusInternalServerError, "Failed to repository client")
		return
	}
	digest, exist, err := repoClient.ManifestExist(data.Tag)
	if err != nil {
		log.Errorf("Failed to get manifest, error: %v", err)
		isj.RenderError(http.StatusInternalServerError, "Failed to get manifest")
		return
	}
	if !exist {
		log.Errorf("The repository based on request: %+v does not exist", data)
		isj.RenderError(http.StatusNotFound, "")
		return
	}
	//Insert job into DB
	j := models.ScanJob{
		Repository: data.Repo,
		Tag:        data.Tag,
		Digest:     digest,
	}
	jid, err := dao.AddScanJob(j)
	if err != nil {
		log.Errorf("Failed to add scan job to DB, error: %v", err)
		isj.RenderError(http.StatusInternalServerError, "Failed to insert scan job data.")
		return
	}
	log.Debugf("Scan job id: %d", jid)
	sj := job.NewScanJob(jid)
	log.Debugf("Sent job to scheduler, job: %v", sj)
	job.Schedule(sj)
}
