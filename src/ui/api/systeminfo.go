package api

import (
	"net/http"
	"path/filepath"
	"syscall"

	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
)

type SystemInfoApi struct {
	api.BaseAPI
	currentUserID int
	isAdmin       bool
}

const harbor_storage_path = "/harbor_storage"

type SystemInfo struct {
	HarborStorage Storage `json:"harbor_storage"`
}

type Storage struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

var systemInfo SystemInfo = SystemInfo{}

func (sia *SystemInfoApi) Prepare() {
	sia.currentUserID = sia.ValidateUser()

	var err error
	sia.isAdmin, err = dao.IsAdminRole(sia.currentUserID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole:%v", err)
		sia.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

func (sia *SystemInfoApi) GetVolumeInfo() {
	if !sia.isAdmin {
		sia.RenderError(http.StatusForbidden, "User does not have admin role.")
		return
	}
	var stat syscall.Statfs_t
	err := syscall.Statfs(filepath.Join("/", harbor_storage_path), &stat)
	if err != nil {
		log.Errorf("Error occurred in syscall.Statfs: %v", err)
		sia.CustomAbort(http.StatusInternalServerError, "Internal error.")
		return
	}
	storage := Storage{
		Total: stat.Blocks * uint64(stat.Bsize),
		Free:  stat.Bfree * uint64(stat.Bsize),
	}
	systemInfo.HarborStorage = storage
	sia.Data["json"] = systemInfo
	sia.ServeJSON()
}
