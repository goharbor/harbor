// Copyright 2018 The Harbor Authors
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

package scan

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"net/http"
	"os"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/job/impl/utils"
)

// All query the DB and Registry for all image and tags,
// then call Harbor's API to scan each of them.
type All struct {
	registryURL          string
	secret               string
	tokenServiceEndpoint string
	harborAPIEndpoint    string
	coreClient           *http.Client
}

// MaxFails implements the interface in job/Interface
func (sa *All) MaxFails() uint {
	return 1
}

// ShouldRetry implements the interface in job/Interface
func (sa *All) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (sa *All) Validate(params job.Parameters) error {
	if len(params) > 0 {
		return fmt.Errorf("the parms should be empty for scan all job")
	}
	return nil
}

// Run implements the interface in job/Interface
func (sa *All) Run(ctx job.Context, params job.Parameters) error {
	logger := ctx.GetLogger()
	logger.Info("Scanning all the images in the registry")
	err := sa.init(ctx)
	if err != nil {
		logger.Errorf("Failed to initialize the job handler, error: %v", err)
		return err
	}

	repos, err := dao.GetRepositories()
	if err != nil {
		logger.Errorf("Failed to get the list of repositories, error: %v", err)
		return err
	}

	for _, r := range repos {
		repoClient, err := utils.NewRepositoryClientForJobservice(r.Name, sa.registryURL, sa.secret, sa.tokenServiceEndpoint)
		if err != nil {
			logger.Errorf("Failed to get repo client for repo: %s, error: %v", r.Name, err)
			continue
		}
		tags, err := repoClient.ListTag()
		if err != nil {
			logger.Errorf("Failed to get tags for repo: %s, error: %v", r.Name, err)
			continue
		}
		for _, t := range tags {
			logger.Infof("Calling harbor-core API to scan image, %s:%s", r.Name, t)
			resp, err := sa.coreClient.Post(fmt.Sprintf("%s/repositories/%s/tags/%s/scan", sa.harborAPIEndpoint, r.Name, t),
				"application/json",
				bytes.NewReader([]byte("{}")))
			if err != nil {
				logger.Errorf("Failed to trigger image scan, error: %v", err)
			} else {
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logger.Errorf("Failed to read response, error: %v", err)
				} else if resp.StatusCode != http.StatusOK {
					logger.Errorf("Unexpected response code: %d, data: %v", resp.StatusCode, data)
				}
				resp.Body.Close()
			}
		}

	}

	return nil
}

func (sa *All) init(ctx job.Context) error {
	if v, err := getAttrFromCtx(ctx, common.RegistryURL); err == nil {
		sa.registryURL = v
	} else {
		return err
	}
	if v := os.Getenv("JOBSERVICE_SECRET"); len(v) > 0 {
		sa.secret = v
	} else {
		return fmt.Errorf("failed to read evnironment variable JOBSERVICE_SECRET")
	}
	sa.coreClient, _ = utils.GetClient()
	if v, err := getAttrFromCtx(ctx, common.TokenServiceURL); err == nil {
		sa.tokenServiceEndpoint = v
	} else {
		return err
	}
	if v, err := getAttrFromCtx(ctx, common.CoreURL); err == nil {
		v = strings.TrimSuffix(v, "/")
		sa.harborAPIEndpoint = v + "/api"
	} else {
		return err
	}
	return nil
}

func getAttrFromCtx(ctx job.Context, key string) (string, error) {
	if v, ok := ctx.Get(key); ok && len(v.(string)) > 0 {
		return v.(string), nil
	}
	return "", fmt.Errorf("failed to get required property: %s", key)
}
