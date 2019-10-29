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
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/systeminfo"
	"github.com/goharbor/harbor/src/core/systeminfo/imagestorage"
	"github.com/goharbor/harbor/src/pkg/version"
)

// SystemInfoAPI handle requests for getting system info /api/systeminfo
type SystemInfoAPI struct {
	BaseController
}

const defaultRootCert = "/etc/core/ca/ca.crt"

// SystemInfo models for system info.
type SystemInfo struct {
	HarborStorage Storage `json:"storage"`
}

// Storage models for storage.
type Storage struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

// GeneralInfo wraps common systeminfo for anonymous request
type GeneralInfo struct {
	WithNotary                  bool                             `json:"with_notary"`
	WithAdmiral                 bool                             `json:"with_admiral"`
	AdmiralEndpoint             string                           `json:"admiral_endpoint"`
	AuthMode                    string                           `json:"auth_mode"`
	AuthProxySettings           *models.HTTPAuthProxy            `json:"authproxy_settings,omitempty"`
	RegistryURL                 string                           `json:"registry_url"`
	ExtURL                      string                           `json:"external_url"`
	ProjectCreationRestrict     string                           `json:"project_creation_restriction"`
	SelfRegistration            bool                             `json:"self_registration"`
	HasCARoot                   bool                             `json:"has_ca_root"`
	HarborVersion               string                           `json:"harbor_version"`
	ClairVulnStatus             *models.ClairVulnerabilityStatus `json:"clair_vulnerability_status,omitempty"`
	RegistryStorageProviderName string                           `json:"registry_storage_provider_name"`
	ReadOnly                    bool                             `json:"read_only"`
	WithChartMuseum             bool                             `json:"with_chartmuseum"`
	NotificationEnable          bool                             `json:"notification_enable"`
}

// GetVolumeInfo gets specific volume storage info.
func (sia *SystemInfoAPI) GetVolumeInfo() {
	if !sia.SecurityCtx.IsAuthenticated() {
		sia.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	if !sia.SecurityCtx.IsSysAdmin() {
		sia.SendForbiddenError(errors.New(sia.SecurityCtx.GetUsername()))
		return
	}

	systeminfo.Init()
	capacity, err := imagestorage.GlobalDriver.Cap()
	if err != nil {
		log.Errorf("failed to get capacity: %v", err)
		sia.SendInternalServerError(fmt.Errorf("failed to get capacity: %v", err))
		return
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

// GetCert gets default self-signed certificate.
func (sia *SystemInfoAPI) GetCert() {
	if _, err := os.Stat(defaultRootCert); err == nil {
		sia.Ctx.Output.Header("Content-Type", "application/octet-stream")
		sia.Ctx.Output.Header("Content-Disposition", "attachment; filename=ca.crt")
		http.ServeFile(sia.Ctx.ResponseWriter, sia.Ctx.Request, defaultRootCert)
	} else if os.IsNotExist(err) {
		log.Error("No certificate found.")
		sia.SendNotFoundError(errors.New("no certificate found"))
		return
	} else {
		log.Errorf("Unexpected error: %v", err)
		sia.SendInternalServerError(fmt.Errorf("unexpected error: %v", err))
		return
	}
}

// GetGeneralInfo returns the general system info, which is to be called by anonymous user
func (sia *SystemInfoAPI) GetGeneralInfo() {
	cfg, err := config.GetSystemCfg()
	if err != nil {
		log.Errorf("Error occurred getting config: %v", err)
		sia.SendInternalServerError(fmt.Errorf("unexpected error: %v", err))
		return
	}
	extURL := cfg[common.ExtEndpoint].(string)
	var registryURL string
	if l := strings.Split(extURL, "://"); len(l) > 1 {
		registryURL = l[1]
	} else {
		registryURL = l[0]
	}
	_, caStatErr := os.Stat(defaultRootCert)
	enableCADownload := caStatErr == nil && strings.HasPrefix(extURL, "https://")
	harborVersion := sia.getVersion()
	info := GeneralInfo{
		AdmiralEndpoint:             utils.SafeCastString(cfg[common.AdmiralEndpoint]),
		WithAdmiral:                 config.WithAdmiral(),
		WithNotary:                  config.WithNotary(),
		AuthMode:                    utils.SafeCastString(cfg[common.AUTHMode]),
		ProjectCreationRestrict:     utils.SafeCastString(cfg[common.ProjectCreationRestriction]),
		SelfRegistration:            utils.SafeCastBool(cfg[common.SelfRegistration]),
		ExtURL:                      extURL,
		RegistryURL:                 registryURL,
		HasCARoot:                   enableCADownload,
		HarborVersion:               harborVersion,
		RegistryStorageProviderName: utils.SafeCastString(cfg[common.RegistryStorageProviderName]),
		ReadOnly:                    config.ReadOnly(),
		WithChartMuseum:             config.WithChartMuseum(),
		NotificationEnable:          utils.SafeCastBool(cfg[common.NotificationEnable]),
	}

	if info.AuthMode == common.HTTPAuth {
		if s, err := config.HTTPAuthProxySetting(); err == nil {
			info.AuthProxySettings = s
		} else {
			log.Warningf("Failed to get auth proxy setting, error: %v", err)
		}
	}
	sia.Data["json"] = info
	sia.ServeJSON()
}

// getVersion gets harbor version.
func (sia *SystemInfoAPI) getVersion() string {
	return fmt.Sprintf("%s-%s", version.ReleaseVersion, version.GitCommit)
}

// Ping ping the harbor core service.
func (sia *SystemInfoAPI) Ping() {
	sia.Data["json"] = "Pong"
	sia.ServeJSON()
}
