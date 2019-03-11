package api

import (
	"net/http"
	"strconv"

	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/core/config"
)

// ScanAllAPI handles request of harbor admin...
type ScanAllAPI struct {
	AJAPI
}

// Prepare validates the URL and parms, it needs the system admin permission.
func (sc *ScanAllAPI) Prepare() {
	sc.BaseController.Prepare()
	if !config.WithClair() {
		log.Warningf("Harbor is not deployed with Clair, it's not possible to scan images.")
		sc.RenderError(http.StatusServiceUnavailable, "")
		return
	}
	if !sc.SecurityCtx.IsAuthenticated() {
		sc.HandleUnauthorized()
		return
	}
	if !sc.SecurityCtx.IsSysAdmin() {
		sc.HandleForbidden(sc.SecurityCtx.GetUsername())
		return
	}
}

// Post ...
func (sc *ScanAllAPI) Post() {
	ajr := models.AdminJobReq{}
	sc.DecodeJSONReqAndValidate(&ajr)
	ajr.Name = common_job.ImageScanAllJob
	sc.submitAdminJob(&ajr)
	sc.Redirect(http.StatusCreated, strconv.FormatInt(ajr.ID, 10))
}

// Put ...
func (sc *ScanAllAPI) Put() {
	ajr := models.AdminJobReq{}
	sc.DecodeJSONReqAndValidate(&ajr)
	ajr.Name = common_job.ImageScanAllJob
	sc.updateAdminSchedule(ajr)
}

// Get gets GC schedule ...
func (sc *ScanAllAPI) Get() {
	sc.getAdminSchedule(common_job.ImageScanAllJob)
}
