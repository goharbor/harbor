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
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

//SystemInfoAPI handle requests for getting system info /api/systeminfo
type SystemInfoAPI struct {
	api.BaseAPI
	currentUserID int
	isAdmin       bool
}

const defaultRootCert = "/etc/ui/ca/ca.crt"
const harborVersionFile = "/harbor/VERSION"

//SystemInfo models for system info.
type SystemInfo struct {
	HarborStorage Storage `json:"storage"`
}

//Storage models for storage.
type Storage struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

//GeneralInfo wraps common systeminfo for anonymous request
type GeneralInfo struct {
	WithNotary              bool   `json:"with_notary"`
	WithAdmiral             bool   `json:"with_admiral"`
	AdmiralEndpoint         string `json:"admiral_endpoint"`
	AuthMode                string `json:"auth_mode"`
	RegistryURL             string `json:"registry_url"`
	ProjectCreationRestrict string `json:"project_creation_restriction"`
	SelfRegistration        bool   `json:"self_registration"`
	HasCARoot               bool   `json:"has_ca_root"`
	HarborVersion           string `json:"harbor_version"`
}

// validate for validating user if an admin.
func (sia *SystemInfoAPI) validate() {
	sia.currentUserID = sia.ValidateUser()

	var err error
	sia.isAdmin, err = dao.IsAdminRole(sia.currentUserID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole:%v", err)
		sia.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

// GetVolumeInfo gets specific volume storage info.
func (sia *SystemInfoAPI) GetVolumeInfo() {
	sia.validate()
	if !sia.isAdmin {
		sia.RenderError(http.StatusForbidden, "User does not have admin role.")
		return
	}

	capacity, err := config.AdminserverClient.Capacity()
	if err != nil {
		log.Errorf("failed to get capacity: %v", err)
		sia.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	systemInfo := SystemInfo{
		HarborStorage: Storage{
			Total: capacity.Total,
			Free:  capacity.Free,
		},
	}

	sia.Data["json"] = systemInfo
	sia.ServeJSON()
}

//GetCert gets default self-signed certificate.
func (sia *SystemInfoAPI) GetCert() {
	sia.validate()
	if sia.isAdmin {
		if _, err := os.Stat(defaultRootCert); err == nil {
			sia.Ctx.Output.Header("Content-Type", "application/octet-stream")
			sia.Ctx.Output.Header("Content-Disposition", "attachment; filename=ca.crt")
			http.ServeFile(sia.Ctx.ResponseWriter, sia.Ctx.Request, defaultRootCert)
		} else if os.IsNotExist(err) {
			log.Error("No certificate found.")
			sia.CustomAbort(http.StatusNotFound, "No certificate found.")
		} else {
			log.Errorf("Unexpected error: %v", err)
			sia.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}
	sia.CustomAbort(http.StatusForbidden, "")
}

// GetGeneralInfo returns the general system info, which is to be called by anonymous user
func (sia *SystemInfoAPI) GetGeneralInfo() {
	cfg, err := config.GetSystemCfg()
	if err != nil {
		log.Errorf("Error occured getting config: %v", err)
		sia.CustomAbort(http.StatusInternalServerError, "Unexpected error")
	}
	var registryURL string
	if l := strings.Split(cfg[common.ExtEndpoint].(string), "://"); len(l) > 1 {
		registryURL = l[1]
	} else {
		registryURL = l[0]
	}
	_, caStatErr := os.Stat(defaultRootCert)
	harborVersion := sia.getVersion()
	info := GeneralInfo{
		AdmiralEndpoint:         cfg[common.AdmiralEndpoint].(string),
		WithAdmiral:             config.WithAdmiral(),
		WithNotary:              config.WithNotary(),
		AuthMode:                cfg[common.AUTHMode].(string),
		ProjectCreationRestrict: cfg[common.ProjectCreationRestriction].(string),
		SelfRegistration:        cfg[common.SelfRegistration].(bool),
		RegistryURL:             registryURL,
		HasCARoot:               caStatErr == nil,
		HarborVersion:           harborVersion,
	}
	sia.Data["json"] = info
	sia.ServeJSON()
}

// GetVersion gets harbor version.
func (sia *SystemInfoAPI) getVersion() string {
	version, err := ioutil.ReadFile(harborVersionFile)
	if err != nil {
		log.Errorf("Error occured getting harbor version: %v", err)
		return ""
	}
	return string(version[:])
}
