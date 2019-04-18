package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/job/models"
)

const (
	jobUUID    = "u-1234-5678-9012"
	jobsPrefix = "/api/v1/jobs"
)

func currPath() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	return path.Dir(f)
}

// NewJobServiceServer ...
func NewJobServiceServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("%s/%s/log", jobsPrefix, jobUUID),
		func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodGet {
				rw.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusOK)
			f := path.Join(currPath(), "test.log")
			b, _ := ioutil.ReadFile(f)
			_, err := rw.Write(b)
			if err != nil {
				panic(err)
			}
		})
	mux.HandleFunc(fmt.Sprintf("%s/%s", jobsPrefix, jobUUID),
		func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodPost {
				rw.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}
			action := models.JobActionRequest{}
			if err := json.Unmarshal(data, &action); err != nil {
				panic(err)
			}
			if strings.ToLower(action.Action) != "stop" && strings.ToLower(action.Action) != "cancel" && strings.ToLower(action.Action) != "retry" {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			rw.WriteHeader(http.StatusNoContent)
			return
		})
	mux.HandleFunc(fmt.Sprintf("%s", jobsPrefix),
		func(rw http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodPost {
				data, err := ioutil.ReadAll(req.Body)
				if err != nil {
					panic(err)
				}
				jobReq := models.JobRequest{}
				json.Unmarshal(data, &jobReq)
				if jobReq.Job.Name == "replication" {
					respData := models.JobStats{
						Stats: &models.JobStatData{
							JobID:    jobUUID,
							Status:   "Pending",
							RunAt:    time.Now().Unix(),
							IsUnique: false,
						},
					}
					b, _ := json.Marshal(respData)
					rw.WriteHeader(http.StatusAccepted)
					if _, err := rw.Write(b); err != nil {
						panic(err)
					}
					return

				}
			}
		})
	return httptest.NewServer(mux)
}
