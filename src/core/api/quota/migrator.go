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

package models

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"strconv"
)

// QuotaMigrator ...
type QuotaMigrator interface {
	// Dump exports all data from backend service, registry, chartmuseum
	Dump() ([]ProjectInfo, error)

	// Usage computes the quota usage of all the projects
	Usage([]ProjectInfo) ([]ProjectUsage, error)

	// Persist record the data to DB, artifact, artifact_blob and blob tabel.
	Persist([]ProjectInfo) error
}

// ProjectInfo ...
type ProjectInfo struct {
	Name  string
	Repos []RepoData
}

// RepoData ...
type RepoData struct {
	Name  string
	Afs   []*models.Artifact
	Afnbs []*models.ArtifactAndBlob
	Blobs []*models.Blob
}

// ProjectUsage ...
type ProjectUsage struct {
	Project string
	Used    quota.ResourceList
}

// Instance ...
type Instance func(promgr.ProjectManager) QuotaMigrator

var adapters = make(map[string]Instance)

// Register ...
func Register(name string, adapter Instance) {
	if adapter == nil {
		panic("quota: Register adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("quota: Register called twice for adapter " + name)
	}
	adapters[name] = adapter
}

// Sync ...
func Sync(pm promgr.ProjectManager, populate bool) error {
	for name := range adapters {
		if !config.WithChartMuseum() {
			if name == "chart" {
				continue
			}
		}
		instanceFunc, ok := adapters[name]
		if !ok {
			err := fmt.Errorf("quota migtator: unknown adapter name %q", name)
			return err
		}
		adapter := instanceFunc(pm)
		data, err := adapter.Dump()
		if err != nil {
			return err
		}
		usage, err := adapter.Usage(data)
		if err := ensureQuota(usage); err != nil {
			return err
		}
		if populate {
			if err := adapter.Persist(data); err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureQuota updates the quota and quota usage in the data base.
func ensureQuota(usages []ProjectUsage) error {
	var pid int64
	for _, usage := range usages {
		project, err := dao.GetProjectByName(usage.Project)
		if err != nil {
			log.Error(err)
			return err
		}
		pid = project.ProjectID
		quotaMgr, err := quota.NewManager("project", strconv.FormatInt(pid, 10))
		if err != nil {
			log.Errorf("Error occurred when to new quota manager %v", err)
			return err
		}
		if err := quotaMgr.EnsureQuota(usage.Used); err != nil {
			log.Errorf("cannot ensure quota for the project: %d, err: %v", pid, err)
			return err
		}
	}
	return nil
}
