package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	common_job "github.com/goharbor/harbor/src/common/job"
	cm "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/scan/all"
	"github.com/goharbor/harbor/src/pkg/scan/api/scan"
	"github.com/goharbor/harbor/src/pkg/scan/api/scanner"
	"github.com/pkg/errors"
)

// ScanAllAPI handles request of scan all images...
type ScanAllAPI struct {
	AJAPI
}

// Prepare validates the URL and parms, it needs the system admin permission.
func (sc *ScanAllAPI) Prepare() {
	sc.BaseController.Prepare()

	if !sc.SecurityCtx.IsAuthenticated() {
		sc.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
	if !sc.SecurityCtx.IsSysAdmin() {
		sc.SendForbiddenError(errors.New(sc.SecurityCtx.GetUsername()))
		return
	}

	enabled, err := isScanEnabled()
	if err != nil {
		sc.SendInternalServerError(err)
		return
	}

	if !enabled {
		sc.SendStatusServiceUnavailableError(errors.New("no scanner is configured, it's not possible to scan"))
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

// GetScheduleMetrics returns the progress metrics for the latest scheduled scan all job
func (sc *ScanAllAPI) GetScheduleMetrics() {
	sc.getMetrics(common_job.JobKindPeriodic)
}

// GetScanAllMetrics returns the progress metrics for the latest manually triggered scan all job
func (sc *ScanAllAPI) GetScanAllMetrics() {
	sc.getMetrics(common_job.JobKindGeneric)
}

func (sc *ScanAllAPI) getMetrics(kind string) {
	aj, err := sc.getLatestAdminJob(common_job.ImageScanAllJob, kind)
	if err != nil {
		sc.SendInternalServerError(errors.Wrap(err, "get metrics: scan all API"))
		return
	}

	var sts *all.Stats
	if aj != nil {
		sts, err = scan.DefaultController.GetStats(fmt.Sprintf("%d", aj.ID))
		if err != nil {
			sc.SendInternalServerError(errors.Wrap(err, "get metrics: scan all API"))
			return
		}

		setOngoing(sts, aj.Status)
	}

	// Return empty
	if sts == nil {
		sts = &all.Stats{}
	}

	sc.Data["json"] = sts
	sc.ServeJSON()
}

func isScanEnabled() (bool, error) {
	kws := make(map[string]interface{})
	kws["is_default"] = true

	query := &q.Query{
		Keywords: kws,
	}

	l, err := scanner.DefaultController.ListRegistrations(query)
	if err != nil {
		return false, errors.Wrap(err, "scan all API: check if scan is enabled")
	}

	return len(l) > 0, nil
}

func setOngoing(stats *all.Stats, st string) {
	status := job.PendingStatus

	if st == cm.JobFinished {
		status = job.SuccessStatus
	} else {
		status = job.Status(strings.ToTitle(st))
	}

	stats.Ongoing = !status.Final() || stats.Total != stats.Completed
}
