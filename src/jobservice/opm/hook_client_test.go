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
package opm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/models"
)

func TestHookClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	err := DefaultHookClient.ReportStatus(ts.URL, models.JobStatusChange{
		JobID:  "fake_job_ID",
		Status: "running",
	})
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

	err := DefaultHookClient.ReportStatus(ts.URL, models.JobStatusChange{
		JobID:  "fake_job_ID",
		Status: "running",
	})
	if err == nil {
		t.Fatal("expect error but got nil")
	}
}
