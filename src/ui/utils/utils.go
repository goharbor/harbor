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
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/service/token"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// RequestAsUI is a shortcut to make a request attach UI secret and send the request.
// Do not use this when you want to handle the response
func RequestAsUI(method, url string, body io.Reader, h ResponseHandler) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	AddUISecret(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	return h.Handle(resp)
}

//AddUISecret add secret cookie to a request
func AddUISecret(req *http.Request) {
	if req != nil {
		req.AddCookie(&http.Cookie{
			Name:  models.UISecretCookie,
			Value: config.UISecret(),
		})
	}
}

// TriggerImageScan triggers an image scan job on jobservice.
func TriggerImageScan(repository string, tag string) error {
	data := &models.ImageScanReq{
		Repo: repository,
		Tag:  tag,
	}
	b, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/jobs/scan", config.InternalJobServiceURL())
	return RequestAsUI("POST", url, bytes.NewBuffer(b), NewStatusRespHandler(http.StatusOK))
}

// NewRepositoryClientForUI creates a repository client that can only be used to
// access the internal registry
func NewRepositoryClientForUI(username, repository string) (*registry.Repository, error) {
	endpoint, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}

	authorizer := auth.NewRawTokenAuthorizer(username, token.Registry)
	transport := registry.NewTransport(http.DefaultTransport, authorizer)
	client := &http.Client{
		Transport: transport,
	}
	return registry.NewRepository(repository, endpoint, client)
}
