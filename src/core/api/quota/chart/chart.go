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

package chart

import (
	"fmt"
	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	common_quota "github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api"
	quota "github.com/goharbor/harbor/src/core/api/quota"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/pkg/errors"
	"net/url"
	"strings"
	"sync"
)

// Migrator ...
type Migrator struct {
	pm promgr.ProjectManager
}

// NewChartMigrator returns a new RegistryMigrator.
func NewChartMigrator(pm promgr.ProjectManager) quota.QuotaMigrator {
	migrator := Migrator{
		pm: pm,
	}
	return &migrator
}

var (
	controller     *chartserver.Controller
	controllerErr  error
	controllerOnce sync.Once
)

// Ping ...
func (rm *Migrator) Ping() error {
	return quota.Check(api.HealthCheckerRegistry["chartmuseum"].Check)
}

// Dump ...
// Depends on DB to dump chart data, as chart cannot get all of namespaces.
func (rm *Migrator) Dump() ([]quota.ProjectInfo, error) {
	var (
		projects []quota.ProjectInfo
		wg       sync.WaitGroup
		err      error
	)

	all, err := dao.GetProjects(nil)
	if err != nil {
		return nil, err
	}

	wg.Add(len(all))
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

	for _, project := range all {
		go func(project *models.Project) {
			defer wg.Done()

			var repos []quota.RepoData
			ctr, err := chartController()
			if err != nil {
				errChan <- err
				return
			}

			chartInfo, err := ctr.ListCharts(project.Name)
			if err != nil {
				errChan <- err
				return
			}

			// repo
			for _, chart := range chartInfo {
				var afs []*models.Artifact
				chartVersions, err := ctr.GetChart(project.Name, chart.Name)
				if err != nil {
					errChan <- err
					continue
				}
				for _, chart := range chartVersions {
					af := &models.Artifact{
						PID:    project.ProjectID,
						Repo:   chart.Name,
						Tag:    chart.Version,
						Digest: chart.Digest,
						Kind:   "Chart",
					}
					afs = append(afs, af)
				}
				repoData := quota.RepoData{
					Name: project.Name,
					Afs:  afs,
				}
				repos = append(repos, repoData)
			}

			projectInfo := quota.ProjectInfo{
				Name:  project.Name,
				Repos: repos,
			}

			infoChan <- projectInfo
		}(project)
	}

	wg.Wait()
	close(infoChan)

	<-done

	if err != nil {
		return nil, err
	}

	return projects, nil
}

// Usage ...
// Chart will not cover size.
func (rm *Migrator) Usage(projects []quota.ProjectInfo) ([]quota.ProjectUsage, error) {
	var pros []quota.ProjectUsage
	for _, project := range projects {
		var count int64
		// usage count
		for _, repo := range project.Repos {
			count = count + int64(len(repo.Afs))
		}
		proUsage := quota.ProjectUsage{
			Project: project.Name,
			Used: common_quota.ResourceList{
				common_quota.ResourceCount:   count,
				common_quota.ResourceStorage: 0,
			},
		}
		pros = append(pros, proUsage)
	}
	return pros, nil

}

// Persist ...
// Chart will not persist data into db.
func (rm *Migrator) Persist(projects []quota.ProjectInfo) error {
	return nil
}

func chartController() (*chartserver.Controller, error) {
	controllerOnce.Do(func() {
		addr, err := config.GetChartMuseumEndpoint()
		if err != nil {
			controllerErr = fmt.Errorf("failed to get the endpoint URL of chart storage server: %s", err.Error())
			return
		}

		addr = strings.TrimSuffix(addr, "/")
		url, err := url.Parse(addr)
		if err != nil {
			controllerErr = errors.New("endpoint URL of chart storage server is malformed")
			return
		}

		ctr, err := chartserver.NewController(url)
		if err != nil {
			controllerErr = errors.New("failed to initialize chart API controller")
		}

		controller = ctr

		log.Debugf("Chart storage server is set to %s", url.String())
		log.Info("API controller for chart repository server is successfully initialized")
	})

	return controller, controllerErr
}

func init() {
	quota.Register("chart", NewChartMigrator)
}
