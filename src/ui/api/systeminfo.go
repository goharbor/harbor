package api

import (
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
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

// Prepare for validating user if an admin.
func (sia *SystemInfoAPI) Prepare() {
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
	if sia.isAdmin {
		if _, err := os.Stat(defaultRootCert); !os.IsNotExist(err) {
			sia.Ctx.Output.Header("Content-Disposition", "attachment; filename=ca.crt")
			http.ServeFile(sia.Ctx.ResponseWriter, sia.Ctx.Request, defaultRootCert)
		} else {
			log.Error("No certificate found.")
			sia.CustomAbort(http.StatusNotFound, "No certificate found.")
		}
	}
	sia.CustomAbort(http.StatusUnauthorized, "")
}
