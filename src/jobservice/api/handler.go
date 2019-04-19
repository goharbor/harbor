// Copyright Project Harbor Authors
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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	"fmt"
	"github.com/goharbor/harbor/src/jobservice/common/query"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/core"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/pkg/errors"
	"strconv"
)

const totalHeaderKey = "Total-Count"

// Handler defines approaches to handle the http requests.
type Handler interface {
	// HandleLaunchJobReq is used to handle the job submission request.
	HandleLaunchJobReq(w http.ResponseWriter, req *http.Request)

	// HandleGetJobReq is used to handle the job stats query request.
	HandleGetJobReq(w http.ResponseWriter, req *http.Request)

	// HandleJobActionReq is used to handle the job action requests (stop/retry).
	HandleJobActionReq(w http.ResponseWriter, req *http.Request)

	// HandleCheckStatusReq is used to handle the job service healthy status checking request.
	HandleCheckStatusReq(w http.ResponseWriter, req *http.Request)

	// HandleJobLogReq is used to handle the request of getting job logs
	HandleJobLogReq(w http.ResponseWriter, req *http.Request)

	// HandleJobLogReq is used to handle the request of getting periodic executions
	HandlePeriodicExecutions(w http.ResponseWriter, req *http.Request)

	// HandleScheduledJobs is used to handle the request of getting pending scheduled jobs
	HandleScheduledJobs(w http.ResponseWriter, req *http.Request)
}

// DefaultHandler is the default request handler which implements the Handler interface.
type DefaultHandler struct {
	controller core.Interface
}

// NewDefaultHandler is constructor of DefaultHandler.
func NewDefaultHandler(ctl core.Interface) *DefaultHandler {
	return &DefaultHandler{
		controller: ctl,
	}
}

// HandleLaunchJobReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleLaunchJobReq(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.ReadRequestBodyError(err))
		return
	}

	// unmarshal data
	jobReq := &job.Request{}
	if err = json.Unmarshal(data, jobReq); err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.HandleJSONDataError(err))
		return
	}

	// Pass request to the controller for the follow-up.
	jobStats, err := dh.controller.LaunchJob(jobReq)
	if err != nil {
		code := http.StatusInternalServerError
		if errs.IsBadRequestError(err) {
			// Bad request
			code = http.StatusBadRequest
		} else if errs.IsConflictError(err) {
			// Conflict error
			code = http.StatusConflict
		} else {
			// General error
			err = errs.LaunchJobError(err)
		}

		dh.handleError(w, req, code, err)
		return
	}

	dh.handleJSONData(w, req, http.StatusAccepted, jobStats)
}

// HandleGetJobReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleGetJobReq(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["job_id"]

	jobStats, err := dh.controller.GetJob(jobID)
	if err != nil {
		code := http.StatusInternalServerError
		if errs.IsObjectNotFoundError(err) {
			code = http.StatusNotFound
		} else if errs.IsBadRequestError(err) {
			code = http.StatusBadRequest
		} else {
			err = errs.GetJobStatsError(err)
		}
		dh.handleError(w, req, code, err)
		return
	}

	dh.handleJSONData(w, req, http.StatusOK, jobStats)
}

// HandleJobActionReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleJobActionReq(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["job_id"]

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.ReadRequestBodyError(err))
		return
	}

	// unmarshal data
	jobActionReq := &job.ActionRequest{}
	if err = json.Unmarshal(data, jobActionReq); err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.HandleJSONDataError(err))
		return
	}

	// Only support stop command now
	cmd := job.OPCommand(jobActionReq.Action)
	if !cmd.IsStop() {
		dh.handleError(w, req, http.StatusNotImplemented, errs.UnknownActionNameError(errors.Errorf("command: %s", jobActionReq.Action)))
		return
	}

	// Stop job
	if err := dh.controller.StopJob(jobID); err != nil {
		code := http.StatusInternalServerError
		if errs.IsObjectNotFoundError(err) {
			code = http.StatusNotFound
		} else if errs.IsBadRequestError(err) {
			code = http.StatusBadRequest
		} else {
			err = errs.StopJobError(err)
		}
		dh.handleError(w, req, code, err)
		return
	}

	dh.log(req, http.StatusNoContent, string(data))

	w.WriteHeader(http.StatusNoContent) // only header, no content returned
}

// HandleCheckStatusReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleCheckStatusReq(w http.ResponseWriter, req *http.Request) {
	stats, err := dh.controller.CheckStatus()
	if err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.CheckStatsError(err))
		return
	}

	dh.handleJSONData(w, req, http.StatusOK, stats)
}

// HandleJobLogReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleJobLogReq(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["job_id"]

	if strings.Contains(jobID, "..") || strings.ContainsRune(jobID, os.PathSeparator) {
		dh.handleError(w, req, http.StatusBadRequest, errors.Errorf("invalid Job ID: %s", jobID))
		return
	}

	logData, err := dh.controller.GetJobLogData(jobID)
	if err != nil {
		code := http.StatusInternalServerError
		if errs.IsObjectNotFoundError(err) {
			code = http.StatusNotFound
		} else if errs.IsBadRequestError(err) {
			code = http.StatusBadRequest
		} else {
			err = errs.GetJobLogError(err)
		}
		dh.handleError(w, req, code, err)
		return
	}

	dh.log(req, http.StatusOK, "")

	w.WriteHeader(http.StatusOK)
	w.Write(logData)
}

// HandleJobLogReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandlePeriodicExecutions(w http.ResponseWriter, req *http.Request) {
	// Get param
	vars := mux.Vars(req)
	jobID := vars["job_id"]

	// Get query params
	q := extractQuery(req)

	executions, total, err := dh.controller.GetPeriodicExecutions(jobID, q)
	if err != nil {
		code := http.StatusInternalServerError
		if errs.IsObjectNotFoundError(err) {
			code = http.StatusNotFound
		} else if errs.IsBadRequestError(err) {
			code = http.StatusBadRequest
		} else {
			err = errs.GetPeriodicExecutionError(err)
		}
		dh.handleError(w, req, code, err)
		return
	}

	w.Header().Add(totalHeaderKey, fmt.Sprintf("%d", total))
	dh.handleJSONData(w, req, http.StatusOK, executions)

}

// HandleScheduledJobs is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleScheduledJobs(w http.ResponseWriter, req *http.Request) {
	// Get query parameters
	q := extractQuery(req)
	jobs, total, err := dh.controller.ScheduledJobs(q)
	if err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.GetScheduledJobsError(err))
		return
	}

	w.Header().Add(totalHeaderKey, fmt.Sprintf("%d", total))
	dh.handleJSONData(w, req, http.StatusOK, jobs)
}

func (dh *DefaultHandler) handleJSONData(w http.ResponseWriter, req *http.Request, code int, object interface{}) {
	data, err := json.Marshal(object)
	if err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.HandleJSONDataError(err))
		return
	}

	logger.Debugf("Serve http request '%s %s': %d %s", req.Method, req.URL.String(), code, data)

	w.Header().Set(http.CanonicalHeaderKey("Accept"), "application/json")
	w.Header().Set(http.CanonicalHeaderKey("content-type"), "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func (dh *DefaultHandler) handleError(w http.ResponseWriter, req *http.Request, code int, err error) {
	// Log all errors
	logger.Errorf("Serve http request '%s %s' error: %d %s", req.Method, req.URL.String(), code, err.Error())

	w.WriteHeader(code)
	w.Write([]byte(err.Error()))
}

func (dh *DefaultHandler) log(req *http.Request, code int, text string) {
	logger.Debugf("Serve http request '%s %s': %d %s", req.Method, req.URL.String(), code, text)
}

func extractQuery(req *http.Request) *query.Parameter {
	q := &query.Parameter{
		PageNumber: 1,
		PageSize:   query.DefaultPageSize,
		Extras:     make(query.ExtraParameters),
	}

	queries := req.URL.Query()
	// Page number
	p := queries.Get(query.QueryParamKeyPage)
	if !utils.IsEmptyStr(p) {
		if pv, err := strconv.ParseUint(p, 10, 32); err == nil {
			if pv > 1 {
				q.PageNumber = uint(pv)
			}
		}
	}

	// Page number
	size := queries.Get(query.QueryParamKeyPageSize)
	if !utils.IsEmptyStr(size) {
		if pz, err := strconv.ParseUint(size, 10, 32); err == nil {
			if pz > 0 {
				q.PageSize = uint(pz)
			}
		}
	}

	// Extra query parameters
	nonStoppedOnly := queries.Get(query.QueryParamKeyNonStoppedOnly)
	if !utils.IsEmptyStr(nonStoppedOnly) {
		if nonStoppedOnlyV, err := strconv.ParseBool(nonStoppedOnly); err == nil {
			q.Extras.Set(query.ExtraParamKeyNonStoppedOnly, nonStoppedOnlyV)
		}
	}

	return q
}
