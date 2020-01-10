package gc

import (
	"bytes"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/registryctl/api"
	"github.com/goharbor/harbor/src/registryctl/config/registry"
	"net/http"
	"os/exec"
	"time"
)

// NewHandler returns the handler to handler blob request
func NewHandler() http.Handler {
	return &handler{}
}

type handler struct{}

// ServeHTTP ...
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		h.start(w, req)
	default:
		api.HandleForbidden(w, req)
	}
}

// Result ...
type Result struct {
	Status    bool      `json:"status"`
	Msg       string    `json:"msg"`
	StartTime time.Time `json:"starttime"`
	EndTime   time.Time `json:"endtime"`
}

// start ...
func (h *handler) start(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("/bin/bash", "-c", "registry garbage-collect --delete-untagged=true "+registry.RegConf)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	start := time.Now()
	log.Debugf("Start to execute garbage collection...")
	if err := cmd.Run(); err != nil {
		log.Errorf("Fail to execute GC: %v, command err: %s", err, errBuf.String())
		api.HandleInternalServerError(w, r)
		return
	}

	gcr := Result{true, outBuf.String(), start, time.Now()}
	if err := api.WriteJSON(w, r, gcr); err != nil {
		log.Errorf("failed to write response: %v", err)
		return
	}
	log.Debugf("Successful to execute garbage collection...")
}
