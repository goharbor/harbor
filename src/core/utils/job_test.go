package utils

import (
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common/job"
	jobmodels "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
)

type jobDataTestEntry struct {
	input  job.ScanJobParams
	expect jobmodels.JobData
}

func TestBuildScanJobData(t *testing.T) {
	assert := assert.New(t)
	testData := []jobDataTestEntry{
		{input: job.ScanJobParams{
			JobID:      123,
			Digest:     "sha256:abcde",
			Repository: "library/ubuntu",
			Tag:        "latest",
		},
			expect: jobmodels.JobData{
				Name: job.ImageScanJob,
				Parameters: map[string]interface{}{
					"job_int_id": 123,
					"repository": "library/ubuntu",
					"tag":        "latest",
					"digest":     "sha256:abcde",
				},
				Metadata: &jobmodels.JobMetadata{
					JobKind:  job.JobKindGeneric,
					IsUnique: false,
				},
				StatusHook: fmt.Sprintf("%s/service/notifications/jobs/scan/%d", config.InternalCoreURL(), 123),
			},
		},
	}
	for _, d := range testData {
		r, err := buildScanJobData(d.input.JobID, d.input.Repository, d.input.Tag, d.input.Digest)
		assert.Nil(err)
		assert.Equal(d.expect.Name, r.Name)
		//		assert.Equal(d.expect.Parameters, r.Parameters)
		assert.Equal(d.expect.StatusHook, r.StatusHook)
	}
}
