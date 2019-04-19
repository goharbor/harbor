package api

import (
	"errors"
	"net/http"
	"strconv"

	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/core/config"
)

// ScanAllAPI handles request of scan all images...
type ScanAllAPI struct {
	AJAPI
}

// Prepare validates the URL and parms, it needs the system admin permission.
func (sc *ScanAllAPI) Prepare() {
	sc.BaseController.Prepare()
	if !config.WithClair() {
		log.Warningf("Harbor is not deployed with Clair, it's not possible to scan images.")
		sc.SendStatusServiceUnavailableError(errors.New(""))
		return
	}
	if !sc.SecurityCtx.IsAuthenticated() {
		sc.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
	if !sc.SecurityCtx.IsSysAdmin() {
		sc.SendForbiddenError(errors.New(sc.SecurityCtx.GetUsername()))
		return
	}
}

// Post according to the request, it creates a cron schedule or a manual trigger for scan all.
// create a daily schedule for scan all
// 	{
//  "schedule": {
//    "type": "Daily",
//    "cron": "0 0 0 * * *"
//  }
//	}
// create a manual trigger for scan all
// 	{
//  "schedule": {
//    "type": "Manual"
//  }
//	}
func (sc *ScanAllAPI) Post() {
	ajr := models.AdminJobReq{}
	isValid, err := sc.DecodeJSONReqAndValidate(&ajr)
	if !isValid {
		sc.SendBadRequestError(err)
		return
	}
	ajr.Name = common_job.ImageScanAllJob
	sc.submit(&ajr)
	sc.Redirect(http.StatusCreated, strconv.FormatInt(ajr.ID, 10))
}

// Put handles scan all cron schedule update/delete.
// Request: delete the schedule of scan all
// 	{
//  "schedule": {
//    "type": "None",
//    "cron": ""
//  }
//	}
func (sc *ScanAllAPI) Put() {
	ajr := models.AdminJobReq{}
	isValid, err := sc.DecodeJSONReqAndValidate(&ajr)
	if !isValid {
		sc.SendBadRequestError(err)
		return
	}
	ajr.Name = common_job.ImageScanAllJob
	sc.updateSchedule(ajr)
}

// Get gets scan all schedule ...
func (sc *ScanAllAPI) Get() {
	sc.getSchedule(common_job.ImageScanAllJob)
}

// List returns the top 10 executions of scan all which includes manual and cron.
func (sc *ScanAllAPI) List() {
	sc.list(common_job.ImageScanAllJob)
}
