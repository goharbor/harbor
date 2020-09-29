package job

import (
	"fmt"
	"math/rand"

	"github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
)

// MockJobClient ...
type MockJobClient struct {
	JobUUID []string
}

// GetJobLog ...
func (mjc *MockJobClient) GetJobLog(uuid string) ([]byte, error) {
	if uuid == "500" {
		return nil, &http.Error{500, "server side error"}
	}
	if mjc.validUUID(uuid) {
		return []byte("some log"), nil
	}
	return nil, &http.Error{404, "not Found"}
}

// SubmitJob ...
func (mjc *MockJobClient) SubmitJob(data *models.JobData) (string, error) {
	uuid := fmt.Sprintf("u-%d", rand.Int())
	mjc.JobUUID = append(mjc.JobUUID, uuid)
	return uuid, nil
}

// PostAction ...
func (mjc *MockJobClient) PostAction(uuid, action string) error {
	if "500" == uuid {
		return &http.Error{500, "server side error"}
	}
	if !mjc.validUUID(uuid) {
		return &http.Error{404, "not Found"}
	}
	return nil
}

// GetExecutions ...
func (mjc *MockJobClient) GetExecutions(uuid string) ([]job.Stats, error) {
	return nil, nil
}

func (mjc *MockJobClient) validUUID(uuid string) bool {
	for _, u := range mjc.JobUUID {
		if uuid == u {
			return true
		}
	}
	return false
}
