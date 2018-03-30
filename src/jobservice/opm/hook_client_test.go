// Copyright 2018 The Harbor Authors. All rights reserved.
package opm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmware/harbor/src/jobservice/models"
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
