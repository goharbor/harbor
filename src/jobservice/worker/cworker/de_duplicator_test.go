package cworker

import (
	"testing"

	"github.com/goharbor/harbor/src/jobservice/tests"
)

func TestDeDuplicator(t *testing.T) {
	jobName := "fake_job"
	jobParams := map[string]interface{}{
		"image": "ubuntu:latest",
	}

	rdd := NewDeDuplicator(tests.GiveMeTestNamespace(), rPool)

	if err := rdd.MustUnique(jobName, jobParams); err != nil {
		t.Error(err)
	}

	if err := rdd.DelUniqueSign(jobName, jobParams); err == nil {
		t.Errorf("expect duplicated error but got nil error")
	}

	if err := rdd.DelUniqueSign(jobName, jobParams); err != nil {
		t.Error(err)
	}
}
