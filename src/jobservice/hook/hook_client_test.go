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
package hook

import (
	"fmt"
	"github.com/goharbor/harbor/src/jobservice/job"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var testClient = NewClient()

func TestHookClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	changeData := &job.StatusChange{
		JobID:  "fake_job_ID",
		Status: "running",
	}
	evt := &Event{
		URL:       ts.URL,
		Data:      changeData,
		Message:   fmt.Sprintf("Status of job %s changed to: %s", changeData.JobID, changeData.Status),
		Timestamp: time.Now().Unix(),
	}
	err := testClient.SendEvent(evt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReportStatusFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed"))
	}))
	defer ts.Close()

	changeData := &job.StatusChange{
		JobID:  "fake_job_ID",
		Status: "running",
	}
	evt := &Event{
		URL:       ts.URL,
		Data:      changeData,
		Message:   fmt.Sprintf("Status of job %s changed to: %s", changeData.JobID, changeData.Status),
		Timestamp: time.Now().Unix(),
	}

	err := testClient.SendEvent(evt)
	if err == nil {
		t.Fatal("expect error but got nil")
	}
}
