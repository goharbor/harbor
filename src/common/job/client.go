package job

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
)

// Client wraps interface to access jobservice.
type Client interface {
	SubmitJob(*models.JobData) (string, error)
	GetJobLog(uuid string) ([]byte, error)
	PostAction(uuid, action string) error
	GetExecutions(uuid string) ([]job.Stats, error)
	// TODO Redirect joblog when we see there's memory issue.
}

// DefaultClient is the default implementation of Client interface
type DefaultClient struct {
	endpoint string
	client   *commonhttp.Client
}

// NewDefaultClient creates a default client based on endpoint and secret.
func NewDefaultClient(endpoint, secret string) *DefaultClient {
	var c *commonhttp.Client
	if len(secret) > 0 {
		c = commonhttp.NewClient(nil, auth.NewSecretAuthorizer(secret))
	} else {
		c = commonhttp.NewClient(nil)
	}
	e := strings.TrimRight(endpoint, "/")
	return &DefaultClient{
		endpoint: e,
		client:   c,
	}
}

// SubmitJob call jobserivce API to submit a job and returns the job's UUID.
func (d *DefaultClient) SubmitJob(jd *models.JobData) (string, error) {
	url := d.endpoint + "/api/v1/jobs"
	jq := models.JobRequest{
		Job: jd,
	}
	b, err := json.Marshal(jq)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusAccepted {
		return "", &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}
	stats := &models.JobStats{}
	if err := json.Unmarshal(data, stats); err != nil {
		return "", err
	}
	return stats.Stats.JobID, nil
}

// GetJobLog call jobserivce API to get the log of a job.  It only accepts the UUID of the job
func (d *DefaultClient) GetJobLog(uuid string) ([]byte, error) {
	url := d.endpoint + "/api/v1/jobs/" + uuid + "/log"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}
	return data, nil
}

// GetExecutions ...
func (d *DefaultClient) GetExecutions(periodicJobID string) ([]job.Stats, error) {
	url := fmt.Sprintf("%s/api/v1/jobs/%s/executions", d.endpoint, periodicJobID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}
	var exes []job.Stats
	err = json.Unmarshal(data, &exes)
	if err != nil {
		return nil, err
	}
	return exes, nil
}

// PostAction call jobservice's API to operate action for job specified by uuid
func (d *DefaultClient) PostAction(uuid, action string) error {
	url := d.endpoint + "/api/v1/jobs/" + uuid
	req := struct {
		Action string `json:"action"`
	}{
		Action: action,
	}
	return d.client.Post(url, req)
}
