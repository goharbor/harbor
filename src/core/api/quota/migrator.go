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
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/types"
	"strconv"
)

// QuotaMigrator ...
type QuotaMigrator interface {
	// Ping validates and wait for backend service ready.
	Ping() error

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
	totalUsage := make(map[string][]ProjectUsage)
	for name, instanceFunc := range adapters {
		if !config.WithChartMuseum() {
			if name == "chart" {
				continue
			}
		}
		adapter := instanceFunc(pm)
		if err := adapter.Ping(); err != nil {
			return err
		}
		data, err := adapter.Dump()
		if err != nil {
			return err
		}
		usage, err := adapter.Usage(data)
		if err != nil {
			return err
		}
		totalUsage[name] = usage
		if populate {
			if err := adapter.Persist(data); err != nil {
				return err
			}
		}
	}
	merged := mergeUsage(totalUsage)
	if err := ensureQuota(merged); err != nil {
		return err
	}
	return nil
}

// mergeUsage merges the usage of adapters
func mergeUsage(total map[string][]ProjectUsage) []ProjectUsage {
	if !config.WithChartMuseum() {
		return total["registry"]
	}
	regUsgs := total["registry"]
	chartUsgs := total["chart"]

	var mergedUsage []ProjectUsage
	temp := make(map[string]quota.ResourceList)

	for _, regUsg := range regUsgs {
		_, exist := temp[regUsg.Project]
		if !exist {
			temp[regUsg.Project] = regUsg.Used
			mergedUsage = append(mergedUsage, ProjectUsage{
				Project: regUsg.Project,
				Used:    regUsg.Used,
			})
		}
	}
	for _, chartUsg := range chartUsgs {
		var usedTemp quota.ResourceList
		_, exist := temp[chartUsg.Project]
		if !exist {
			usedTemp = chartUsg.Used
		} else {
			usedTemp = types.Add(temp[chartUsg.Project], chartUsg.Used)
		}
		temp[chartUsg.Project] = usedTemp
		mergedUsage = append(mergedUsage, ProjectUsage{
			Project: chartUsg.Project,
			Used:    usedTemp,
		})
	}
	return mergedUsage
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
