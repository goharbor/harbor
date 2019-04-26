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
	"github.com/goharbor/harbor/src/jobservice/job"
	job_models "github.com/goharbor/harbor/src/jobservice/job"
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
	mux.HandleFunc(fmt.Sprintf("%s/%s/executions", jobsPrefix, jobUUID),
		func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodGet {
				rw.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var stats []job.Stats
			stat := job_models.Stats{
				Info: &job_models.StatsInfo{
					JobID:    jobUUID,
					Status:   "Pending",
					RunAt:    time.Now().Unix(),
					IsUnique: false,
				},
			}
			stats = append(stats, stat)
			b, _ := json.Marshal(stats)
			if _, err := rw.Write(b); err != nil {
				panic(err)
			}
			rw.WriteHeader(http.StatusOK)
			return
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
						Stats: &models.StatsInfo{
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
