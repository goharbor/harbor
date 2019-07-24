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
	"errors"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"net/http"

	"github.com/ghodss/yaml"
	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/filter"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
)

const (
	yamlFileContentType = "application/x-yaml"
	// ReplicationJobType ...
	ReplicationJobType = "replication"
	// ScanJobType ...
	ScanJobType = "scan"
)

// the managers/controllers used globally
var (
	projectMgr          project.Manager
	repositoryMgr       repository.Manager
	retentionScheduler  scheduler.Scheduler
	retentionMgr        retention.Manager
	retentionLauncher   retention.Launcher
	retentionController retention.APIController
)

// BaseController ...
type BaseController struct {
	api.BaseAPI
	// SecurityCtx is the security context used to authN &authZ
	SecurityCtx security.Context
	// ProjectMgr is the project manager which abstracts the operations
	// related to projects
	ProjectMgr promgr.ProjectManager
}

// Prepare inits security context and project manager from request
// context
func (b *BaseController) Prepare() {
	ctx, err := filter.GetSecurityContext(b.Ctx.Request)
	if err != nil {
		log.Errorf("failed to get security context: %v", err)
		b.SendInternalServerError(errors.New(""))
		return
	}
	b.SecurityCtx = ctx

	pm, err := filter.GetProjectManager(b.Ctx.Request)
	if err != nil {
		log.Errorf("failed to get project manager: %v", err)
		b.SendInternalServerError(errors.New(""))
		return
	}
	b.ProjectMgr = pm
}

// WriteJSONData writes the JSON data to the client.
func (b *BaseController) WriteJSONData(object interface{}) {
	b.Data["json"] = object
	b.ServeJSON()
}

// WriteYamlData writes the yaml data to the client.
func (b *BaseController) WriteYamlData(object interface{}) {
	yData, err := yaml.Marshal(object)
	if err != nil {
		b.SendInternalServerError(err)
		return
	}

	w := b.Ctx.ResponseWriter
	w.Header().Set("Content-Type", yamlFileContentType)
	w.WriteHeader(http.StatusOK)
	w.Write(yData)
}

// Init related objects/configurations for the API controllers
func Init() error {
	registerHealthCheckers()

	// init chart controller
	if err := initChartController(); err != nil {
		return err
	}

	// init project manager
	initProjectManager()

	// init repository manager
	initRepositoryManager()

	initRetentionScheduler()

	retentionMgr = retention.NewManager()

	retentionLauncher = retention.NewLauncher(projectMgr, repositoryMgr, retentionMgr, dep.DefaultClient)

	retentionController = retention.NewAPIController(projectMgr, repositoryMgr, retentionScheduler, retentionLauncher)

	callbackFun := func(p interface{}) error {
		r, ok := p.(retention.TriggerParam)
		if ok {
			return retentionController.TriggerRetentionExec(r.PolicyID, r.Trigger, false)
		}
		return errors.New("bad retention callback param")
	}
	err := scheduler.Register(retention.RetentionSchedulerCallback, callbackFun)

	return err
}

func initChartController() error {
	// If chart repository is not enabled then directly return
	if !config.WithChartMuseum() {
		return nil
	}

	chartCtl, err := initializeChartController()
	if err != nil {
		return err
	}

	chartController = chartCtl
	return nil
}

func initProjectManager() {
	projectMgr = project.New()
}

func initRepositoryManager() {
	repositoryMgr = repository.New(projectMgr, chartController)
}

func initRetentionScheduler() {
	retentionScheduler = scheduler.GlobalScheduler
}
