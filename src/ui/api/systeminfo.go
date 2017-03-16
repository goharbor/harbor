package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/vmware/harbor/src/common/api"
	comcfg "github.com/vmware/harbor/src/common/config"
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

const harborStoragePath = "/harbor_storage"
const defaultRootCert = "/harbor_storage/ca_download/ca.crt"

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
	var stat syscall.Statfs_t
	err := syscall.Statfs(filepath.Join("/", harborStoragePath), &stat)
	if err != nil {
		log.Errorf("Error occurred in syscall.Statfs: %v", err)
		sia.CustomAbort(http.StatusInternalServerError, "Internal error.")
		return
	}

	systemInfo := SystemInfo{
		HarborStorage: Storage{
			Total: stat.Blocks * uint64(stat.Bsize),
			Free:  stat.Bavail * uint64(stat.Bsize),
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
	if l := strings.Split(cfg[comcfg.ExtEndpoint].(string), "://"); len(l) > 1 {
		registryURL = l[1]
	} else {
		registryURL = l[0]
	}
	_, caStatErr := os.Stat(defaultRootCert)
	info := GeneralInfo{
		AdmiralEndpoint:         cfg[comcfg.AdmiralEndpoint].(string),
		WithAdmiral:             config.WithAdmiral(),
		WithNotary:              config.WithNotary(),
		AuthMode:                cfg[comcfg.AUTHMode].(string),
		ProjectCreationRestrict: cfg[comcfg.ProjectCreationRestriction].(string),
		SelfRegistration:        cfg[comcfg.SelfRegistration].(bool),
		RegistryURL:             registryURL,
		HasCARoot:               caStatErr == nil,
	}
	sia.Data["json"] = info
	sia.ServeJSON()
}
