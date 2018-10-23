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
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
)

// MetadataAPI ...
type MetadataAPI struct {
	BaseController
	metaMgr metamgr.ProjectMetadataManager
	project *models.Project
	name    string
}

// Prepare ...
func (m *MetadataAPI) Prepare() {
	m.BaseController.Prepare()

	m.metaMgr = m.ProjectMgr.GetMetadataManager()

	// the project manager doesn't use a project metadata manager
	if m.metaMgr == nil {
		log.Debug("the project manager doesn't use a project metadata manager")
		m.RenderError(http.StatusMethodNotAllowed, "")
		return
	}

	id, err := m.GetInt64FromPath(":id")
	if err != nil || id <= 0 {
		text := "invalid project ID: "
		if err != nil {
			text += err.Error()
		} else {
			text += fmt.Sprintf("%d", id)
		}
		m.HandleBadRequest(text)
		return
	}

	project, err := m.ProjectMgr.Get(id)
	if err != nil {
		m.ParseAndHandleError(fmt.Sprintf("failed to get project %d", id), err)
		return
	}

	if project == nil {
		m.HandleNotFound(fmt.Sprintf("project %d not found", id))
		return
	}

	m.project = project

	switch m.Ctx.Request.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete:
		if !(m.Ctx.Request.Method == http.MethodGet && project.IsPublic()) {
			if !m.SecurityCtx.IsAuthenticated() {
				m.HandleUnauthorized()
				return
			}
			if !m.SecurityCtx.HasReadPerm(project.ProjectID) {
				m.HandleForbidden(m.SecurityCtx.GetUsername())
				return
			}
		}
	default:
		log.Debugf("%s method not allowed", m.Ctx.Request.Method)
		m.RenderError(http.StatusMethodNotAllowed, "")
		return
	}

	name := m.GetStringFromPath(":name")
	if len(name) > 0 {
		m.name = name
		metas, err := m.metaMgr.Get(project.ProjectID, name)
		if err != nil {
			m.HandleInternalServerError(fmt.Sprintf("failed to get metadata of project %d: %v", project.ProjectID, err))
			return
		}
		if len(metas) == 0 {
			m.HandleNotFound(fmt.Sprintf("metadata %s of project %d not found", name, project.ProjectID))
			return
		}
	}
}

// Get ...
func (m *MetadataAPI) Get() {
	var metas map[string]string
	var err error
	if len(m.name) > 0 {
		metas, err = m.metaMgr.Get(m.project.ProjectID, m.name)
	} else {
		metas, err = m.metaMgr.Get(m.project.ProjectID)
	}

	if err != nil {
		m.HandleInternalServerError(fmt.Sprintf("failed to get metadata %s of project %d: %v", m.name, m.project.ProjectID, err))
		return
	}
	m.Data["json"] = metas
	m.ServeJSON()
}

// Post ...
func (m *MetadataAPI) Post() {
	var metas map[string]string
	m.DecodeJSONReq(&metas)

	ms, err := validateProjectMetadata(metas)
	if err != nil {
		m.HandleBadRequest(err.Error())
		return
	}

	if len(ms) != 1 {
		m.HandleBadRequest("invalid request: has no valid key/value pairs or has more than one valid key/value pairs")
		return
	}

	keys := reflect.ValueOf(ms).MapKeys()
	mts, err := m.metaMgr.Get(m.project.ProjectID, keys[0].String())
	if err != nil {
		m.HandleInternalServerError(fmt.Sprintf("failed to get metadata for project %d: %v", m.project.ProjectID, err))
		return
	}

	if len(mts) != 0 {
		m.HandleConflict()
		return
	}

	if err := m.metaMgr.Add(m.project.ProjectID, ms); err != nil {
		m.HandleInternalServerError(fmt.Sprintf("failed to create metadata for project %d: %v", m.project.ProjectID, err))
		return
	}

	m.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)
}

// Put ...
func (m *MetadataAPI) Put() {
	var metas map[string]string
	m.DecodeJSONReq(&metas)

	meta, exist := metas[m.name]
	if !exist {
		m.HandleBadRequest(fmt.Sprintf("must contains key %s", m.name))
		return
	}

	ms, err := validateProjectMetadata(map[string]string{
		m.name: meta,
	})
	if err != nil {
		m.HandleBadRequest(err.Error())
		return
	}

	if err := m.metaMgr.Update(m.project.ProjectID, map[string]string{
		m.name: ms[m.name],
	}); err != nil {
		m.HandleInternalServerError(fmt.Sprintf("failed to update metadata %s of project %d: %v", m.name, m.project.ProjectID, err))
		return
	}
}

// Delete ...
func (m *MetadataAPI) Delete() {
	if err := m.metaMgr.Delete(m.project.ProjectID, m.name); err != nil {
		m.HandleInternalServerError(fmt.Sprintf("failed to delete metadata %s of project %d: %v", m.name, m.project.ProjectID, err))
		return
	}
}

// validate metas and return a new map which contains the valid key/value pairs only
func validateProjectMetadata(metas map[string]string) (map[string]string, error) {
	if len(metas) == 0 {
		return nil, nil
	}

	boolMetas := []string{
		models.ProMetaPublic,
		models.ProMetaEnableContentTrust,
		models.ProMetaPreventVul,
		models.ProMetaAutoScan}

	for _, boolMeta := range boolMetas {
		value, exist := metas[boolMeta]
		if exist {
			b, err := strconv.ParseBool(value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s to bool: %v", value, err)
			}
			metas[boolMeta] = strconv.FormatBool(b)
		}
	}

	value, exist := metas[models.ProMetaSeverity]
	if exist {
		switch strings.ToLower(value) {
		case models.SeverityHigh, models.SeverityMedium, models.SeverityLow, models.SeverityNone:
			metas[models.ProMetaSeverity] = strings.ToLower(value)
		default:
			return nil, fmt.Errorf("invalid severity %s", value)
		}
	}

	return metas, nil
}
