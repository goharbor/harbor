package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/all"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

// ScanAllAPI handles request of scan all images...
type ScanAllAPI struct {
	BaseController
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

	if ajr.Schedule == nil {
		sc.SendBadRequestError(fmt.Errorf("schedule is required"))
		return
	}

	if ajr.Schedule.Type == models.ScheduleNone {
		return
	}

	if ajr.IsPeriodic() {
		schedule, err := sc.getScanAllSchedule()
		if err != nil {
			sc.SendError(err)
			return
		}

		if schedule != nil {
			err := errors.New("fail to set schedule for scan all as always had one, please delete it firstly then to re-schedule")
			sc.SendPreconditionFailedError(err)
			return
		}

		scheduleID, err := sc.createOrUpdateScanAllSchedule(ajr.Schedule.Type, ajr.Schedule.Cron, nil)
		if err != nil {
			sc.SendError(err)
			return
		}

		sc.Redirect(http.StatusCreated, strconv.FormatInt(scheduleID, 10))
	} else {
		execution, err := sc.getLatestScanAllExecution(task.ExecutionTriggerManual)
		if err != nil {
			sc.SendError(err)
			return
		}

		if execution != nil && execution.IsOnGoing() {
			err := errors.Errorf("a previous scan all job aleady exits, its status is %s", execution.Status)
			sc.SendConflictError(err)
			return
		}

		executionID, err := scan.DefaultController.ScanAll(sc.Context(), task.ExecutionTriggerManual, true)
		if err != nil {
			sc.SendError(err)
			return
		}

		sc.Redirect(http.StatusCreated, strconv.FormatInt(executionID, 10))
	}
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

	if ajr.Schedule.Type == models.ScheduleManual {
		err := fmt.Errorf("fail to update scan all schedule as wrong schedule type: %s", ajr.Schedule.Type)
		sc.SendBadRequestError(err)
		return
	}

	schedule, err := sc.getScanAllSchedule()
	if err != nil {
		sc.SendError(err)
		return
	}

	if ajr.Schedule.Type == models.ScheduleNone {
		if schedule != nil {
			err = scheduler.Sched.UnScheduleByID(sc.Context(), schedule.ID)
		}
	} else {
		_, err = sc.createOrUpdateScanAllSchedule(ajr.Schedule.Type, ajr.Schedule.Cron, schedule)
	}

	if err != nil {
		sc.SendError(err)
	}
}

// Get gets scan all schedule ...
func (sc *ScanAllAPI) Get() {
	result := models.AdminJobRep{}

	schedule, err := sc.getScanAllSchedule()
	if err != nil {
		sc.SendError(err)
		return
	}

	if schedule != nil {
		result.ID = schedule.ID
		result.Status = schedule.Status
		result.CreationTime = schedule.CreationTime
		result.UpdateTime = schedule.UpdateTime
		result.Schedule = &models.ScheduleParam{
			Type: schedule.CRONType,
			Cron: schedule.CRON,
		}
	}

	sc.Data["json"] = result
	sc.ServeJSON()
}

// GetScheduleMetrics returns the progress metrics for the latest scheduled scan all job
func (sc *ScanAllAPI) GetScheduleMetrics() {
	sc.getMetrics(task.ExecutionTriggerSchedule)
}

// GetScanAllMetrics returns the progress metrics for the latest manually triggered scan all job
func (sc *ScanAllAPI) GetScanAllMetrics() {
	sc.getMetrics(task.ExecutionTriggerManual)
}

func (sc *ScanAllAPI) getMetrics(trigger string) {
	execution, err := sc.getLatestScanAllExecution(trigger)
	if err != nil {
		sc.SendError(err)
		return
	}

	sts := &all.Stats{}
	if execution != nil && execution.Metrics != nil {
		metrics := execution.Metrics
		sts.Total = uint(metrics.TaskCount)
		sts.Completed = uint(metrics.SuccessTaskCount)
		sts.Metrics = map[string]uint{
			"Pending": uint(metrics.PendingTaskCount),
			"Running": uint(metrics.RunningTaskCount),
			"Success": uint(metrics.SuccessTaskCount),
			"Error":   uint(metrics.ErrorTaskCount),
			"Stopped": uint(metrics.StoppedTaskCount),
		}
		sts.Ongoing = !job.Status(execution.Status).Final() || sts.Total != sts.Completed
	}

	sc.Data["json"] = sts
	sc.ServeJSON()
}

func (sc *ScanAllAPI) getScanAllSchedule() (*scheduler.Schedule, error) {
	query := q.New(q.KeyWords{"vendor_type": job.ImageScanAllJob})
	schedules, err := scheduler.Sched.ListSchedules(sc.Context(), query.First("-creation_time"))
	if err != nil {
		return nil, err
	}

	if len(schedules) > 1 {
		msg := "found more than one scheduled scan all job, please ensure that only one schedule left"
		return nil, errors.BadRequestError(nil).WithMessage(msg)
	} else if len(schedules) == 0 {
		return nil, nil
	}

	return schedules[0], nil
}

func (sc *ScanAllAPI) createOrUpdateScanAllSchedule(cronType, cron string, previous *scheduler.Schedule) (int64, error) {
	if previous != nil {
		if cronType == previous.CRONType && cron == previous.CRON {
			return previous.ID, nil
		}

		if err := scheduler.Sched.UnScheduleByID(sc.Context(), previous.ID); err != nil {
			return 0, err
		}
	}

	return scheduler.Sched.Schedule(sc.Context(), job.ImageScanAllJob, 0, cronType, cron, scan.ScanAllCallback, nil, nil)
}

func (sc *ScanAllAPI) getLatestScanAllExecution(trigger string) (*task.Execution, error) {
	query := q.New(q.KeyWords{"vendor_type": job.ImageScanAllJob, "trigger": trigger})
	executions, err := task.ExecMgr.List(sc.Context(), query.First("-start_time"))
	if err != nil {
		return nil, err
	}

	if len(executions) == 0 {
		return nil, nil
	}

	return executions[0], nil
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
