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

package replicator

import (
	"fmt"
	"strings"

	"github.com/vmware/harbor/src/common/dao"
	common_job "github.com/vmware/harbor/src/common/job"
	job_models "github.com/vmware/harbor/src/common/job/models"
	common_models "github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/replication/models"
	"github.com/vmware/harbor/src/ui/config"
)

// Replication holds information for a replication
type Replication struct {
	PolicyID   int64
	Candidates []models.FilterItem
	Targets    []*common_models.RepTarget
	Operation  string
}

// Replicator submits the replication work to the jobservice
type Replicator interface {
	Replicate(*Replication) error
}

// DefaultReplicator provides a default implement for Replicator
type DefaultReplicator struct {
	client common_job.Client
}

// NewDefaultReplicator returns an instance of DefaultReplicator
func NewDefaultReplicator(client common_job.Client) *DefaultReplicator {
	return &DefaultReplicator{
		client: client,
	}
}

// Replicate ...
func (d *DefaultReplicator) Replicate(replication *Replication) error {
	repositories := map[string][]string{}
	// TODO the operation of all candidates are same for now. Update it after supporting
	// replicate deletion
	operation := ""
	for _, candidate := range replication.Candidates {
		strs := strings.SplitN(candidate.Value, ":", 2)
		repositories[strs[0]] = append(repositories[strs[0]], strs[1])
		operation = candidate.Operation
	}

	for _, target := range replication.Targets {
		for repository, tags := range repositories {
			// create job in database
			id, err := dao.AddRepJob(common_models.RepJob{
				PolicyID:   replication.PolicyID,
				Repository: repository,
				TagList:    tags,
				Operation:  operation,
			})
			if err != nil {
				return err
			}

			// submit job to jobservice
			log.Debugf("submiting replication job to jobservice, repository: %s, tags: %v, operation: %s, target: %s",
				repository, tags, operation, target.URL)
			job := &job_models.JobData{
				Metadata: &job_models.JobMetadata{
					JobKind: common_job.JobKindGeneric,
				},
				StatusHook: fmt.Sprintf("%s/service/notifications/jobs/replication/%d",
					config.InternalUIURL(), id),
			}

			if operation == common_models.RepOpTransfer {
				url, err := config.ExtEndpoint()
				if err != nil {
					return err
				}
				job.Name = common_job.ImageTransfer
				job.Parameters = map[string]interface{}{
					"repository":            repository,
					"tags":                  tags,
					"src_registry_url":      url,
					"src_registry_insecure": true,
					//"src_token_service_url":"",
					"dst_registry_url":      target.URL,
					"dst_registry_insecure": target.Insecure,
					"dst_registry_username": target.Username,
					"dst_registry_password": target.Password,
				}
			} else {
				job.Name = common_job.ImageDelete
				job.Parameters = map[string]interface{}{
					"repository":            repository,
					"tags":                  tags,
					"dst_registry_url":      target.URL,
					"dst_registry_insecure": target.Insecure,
					"dst_registry_username": target.Username,
					"dst_registry_password": target.Password,
				}
			}

			uuid, err := d.client.SubmitJob(job)
			if err != nil {
				return err
			}

			// create the mapping relationship between the jobs in database and jobservice
			if err = dao.SetRepJobUUID(id, uuid); err != nil {
				return err
			}
		}
	}
	return nil
}
