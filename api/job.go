package api

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
	"net/http"
	"strconv"
)

type JobAPI struct {
	BaseAPI
}

func (ja *JobAPI) Post() {
	var je models.JobEntry
	ja.DecodeJSONReq(&je)
	res, err := json.Marshal(je.Options)
	if !job.RunnerExists(je.Type) {
		log.Errorf("runner for type %s is not registered", je.Type)
		ja.RenderError(http.StatusBadRequest, fmt.Sprintf("runner for type %s is not registered", je.Type))
		return
	}
	je.OptionsStr = string(res)
	if err != nil {
		log.Warningf("Error marshaling options: %v", err)
	}
	res, err = json.Marshal(je.Parms)
	je.ParmsStr = string(res)
	if err != nil {
		log.Warningf("Error marshaling parms: %v", err)
	}
	jobID, err := dao.AddJob(je)
	if err != nil {
		log.Errorf("Failed to add job to DB, error: %v", err)
		ja.RenderError(http.StatusInternalServerError, "Failed to add job")
		return
	}
	je.ID = jobID
	log.Debugf("job Id:%d, type: %s", je.ID, je.Type)
	job.Schedule(je)
}

func (ja *JobAPI) Get() {
	idStr := ja.Ctx.Input.Param(":id")
	if len(idStr) > 0 {
		jobID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Errorf("Failed to parse job id in url: %s", idStr)
			ja.RenderError(http.StatusBadRequest, "invalid job id")
			return
		}
		je, err := dao.GetJob(jobID)
		if err != nil {
			log.Errorf("Failed to query job from db, error: %v", err)
			ja.RenderError(http.StatusInternalServerError, "Failed to query job")
			return
		}
		if je == nil {
			log.Errorf("job does not exist, id: %d", jobID)
			ja.RenderError(http.StatusNotFound, "")
			return
		}
		logs, err := dao.GetJobLogs(jobID)
		if err != nil {
			log.Errorf("Failed to get job logs, error: %v", err)
			ja.RenderError(http.StatusInternalServerError, "Failed to query job")
			return
		}
		je.Logs = logs
		ja.Data["json"] = je
	} else {
		jobs, err := dao.ListJobs()
		if err != nil {
			log.Errorf("Failed to list jobs, error:%v", err)
			ja.RenderError(http.StatusInternalServerError, "Failed to query job")
		}
		log.Debugf("jobs: %v", jobs)
		ja.Data["json"] = jobs
	}
	ja.ServeJSON()
}
