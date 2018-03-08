package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/vmware/harbor/src/jobservice_v2/core"
	"github.com/vmware/harbor/src/jobservice_v2/errs"
	"github.com/vmware/harbor/src/jobservice_v2/models"
)

//Handler defines approaches to handle the http requests.
type Handler interface {
	//HandleLaunchJobReq is used to handle the job submission request.
	HandleLaunchJobReq(w http.ResponseWriter, req *http.Request)

	//HandleGetJobReq is used to handle the job stats query request.
	HandleGetJobReq(w http.ResponseWriter, req *http.Request)

	//HandleJobActionReq is used to handle the job action requests (stop/retry).
	HandleJobActionReq(w http.ResponseWriter, req *http.Request)

	//HandleCheckStatusReq is used to handle the job service healthy status checking request.
	HandleCheckStatusReq(w http.ResponseWriter, req *http.Request)
}

//DefaultHandler is the default request handler which implements the Handler interface.
type DefaultHandler struct {
	context    core.BaseContext
	controller *core.Controller
}

//NewDefaultHandler is constructor of DefaultHandler.
func NewDefaultHandler(ctx core.BaseContext, ctl *core.Controller) *DefaultHandler {
	return &DefaultHandler{
		context:    ctx,
		controller: ctl,
	}
}

//HandleLaunchJobReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleLaunchJobReq(w http.ResponseWriter, req *http.Request) {
	if !dh.preCheck(w) {
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		dh.handleError(w, http.StatusInternalServerError, errs.ReadRequestBodyError(err))
		return
	}

	//unmarshal data
	jobReq := models.JobRequest{}
	if err = json.Unmarshal(data, &jobReq); err != nil {
		dh.handleError(w, http.StatusInternalServerError, errs.HandleJSONDataError(err))
		return
	}

	//Pass request to the controller for the follow-up.
	jobStats, err := dh.controller.LaunchJob(dh.context, jobReq)
	if err != nil {
		dh.handleError(w, http.StatusInternalServerError, errs.LaunchJobError(err))
		return
	}
	data, err = json.Marshal(jobStats)
	if err != nil {
		dh.handleError(w, http.StatusInternalServerError, errs.HandleJSONDataError(err))
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write(data)
}

//HandleGetJobReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleGetJobReq(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["job_id"]
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("job is coming " + jobID))
}

//HandleJobActionReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleJobActionReq(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

//HandleCheckStatusReq is implementation of method defined in interface 'Handler'
func (dh *DefaultHandler) HandleCheckStatusReq(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (dh *DefaultHandler) handleError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte(err.Error()))
}

func (dh *DefaultHandler) preCheck(w http.ResponseWriter) bool {
	if dh.controller == nil {
		dh.handleError(w, http.StatusInternalServerError, errs.MissingBackendHandlerError(fmt.Errorf("nil controller")))
		return false
	}

	return true
}
