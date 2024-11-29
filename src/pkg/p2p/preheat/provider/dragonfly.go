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

package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/client"

	"github.com/olekukonko/tablewriter"
)

const (
	// dragonflyHealthPath is the health check path for dragonfly openapi.
	dragonflyHealthPath = "/healthy"

	// dragonflyJobPath is the job path for dragonfly openapi.
	dragonflyJobPath = "/oapi/v1/jobs"
)

const (
	// dragonflyJobPendingState is the pending state of the job, which means
	// the job is waiting to be processed and running.
	dragonflyJobPendingState = "PENDING"

	// dragonflyJobSuccessState is the success state of the job, which means
	// the job is processed successfully.
	dragonflyJobSuccessState = "SUCCESS"

	// dragonflyJobFailureState is the failure state of the job, which means
	// the job is processed failed.
	dragonflyJobFailureState = "FAILURE"
)

type dragonflyCreateJobRequest struct {
	// Type is the job type, support preheat.
	Type string `json:"type"`

	// Args is the preheating args.
	Args dragonflyCreateJobRequestArgs `json:"args"`

	// SchedulerClusterIDs is the scheduler cluster ids for preheating.
	SchedulerClusterIDs []uint `json:"scheduler_cluster_ids"`
}

type dragonflyCreateJobRequestArgs struct {
	// Type is the preheating type, support image and file.
	Type string `json:"type"`

	// URL is the image url for preheating.
	URL string `json:"url"`

	// Tag is the tag for preheating.
	Tag string `json:"tag"`

	// FilteredQueryParams is the filtered query params for preheating.
	FilteredQueryParams string `json:"filtered_query_params"`

	// Headers is the http headers for authentication.
	Headers map[string]string `json:"headers"`

	// Scope is the scope for preheating, default is single_peer.
	Scope string `json:"scope"`

	// BatchSize is the batch size for preheating all peers, default is 50.
	ConcurrentCount int64 `json:"concurrent_count"`

	// Timeout is the timeout for preheating, default is 30 minutes.
	Timeout time.Duration `json:"timeout"`
}

type dragonflyJobResponse struct {
	// ID is the job id.
	ID int `json:"id"`

	// CreatedAt is the job created time.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the job updated time.
	UpdatedAt time.Time `json:"updated_at"`

	// State is the job state, support PENDING, SUCCESS, FAILURE.
	State string `json:"state"`

	// Results is the job results.
	Result struct {

		// JobStates is the job states, including each job state.
		JobStates []struct {

			// Error is the job error message.
			Error string `json:"error"`

			// Results is the job results.
			Results []struct {

				// SuccessTasks is the success tasks.
				SuccessTasks []*struct {

					// URL is the url of the task, which is the blob url.
					URL string `json:"url"`

					// Hostname is the hostname of the task.
					Hostname string `json:"hostname"`

					// IP is the ip of the task.
					IP string `json:"ip"`
				} `json:"success_tasks"`

				// FailureTasks is the failure tasks.
				FailureTasks []*struct {

					// URL is the url of the task, which is the blob url.
					URL string `json:"url"`

					// Hostname is the hostname of the task.
					Hostname string `json:"hostname"`

					// IP is the ip of the task.
					IP string `json:"ip"`

					// Description is the failure description.
					Description string `json:"description"`
				} `json:"failure_tasks"`

				// SchedulerClusterID is the scheduler cluster id.
				SchedulerClusterID uint `json:"scheduler_cluster_id"`
			} `json:"results"`
		} `json:"job_states"`
	} `json:"result"`
}

// dragonflyExtraAttrs is the extra attributes model definition for dragonfly provider.
type dragonflyExtraAttrs struct {
	// ClusterIDs is the cluster ids for dragonfly provider.
	ClusterIDs []uint `json:"cluster_ids"`
}

// DragonflyDriver implements the provider driver interface for Alibaba dragonfly.
// More details, please refer to https://github.com/alibaba/Dragonfly
type DragonflyDriver struct {
	instance *provider.Instance
}

// Self implements @Driver.Self.
func (dd *DragonflyDriver) Self() *Metadata {
	return &Metadata{
		ID:          "dragonfly",
		Name:        "Dragonfly",
		Icon:        "https://raw.githubusercontent.com/dragonflyoss/Dragonfly2/master/docs/images/logo/dragonfly-linear.png",
		Version:     "2.1.59",
		Source:      "https://github.com/dragonflyoss/Dragonfly2",
		Maintainers: []string{"chlins.zhang@gmail.com", "gaius.qi@gmail.com"},
	}
}

// GetHealth implements @Driver.GetHealth.
func (dd *DragonflyDriver) GetHealth() (*DriverStatus, error) {
	if dd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	url := fmt.Sprintf("%s%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), dragonflyHealthPath)
	url, err := lib.ValidateHTTPURL(url)
	if err != nil {
		return nil, err
	}

	if _, err = client.GetHTTPClient(dd.instance.Insecure).Get(url, dd.getCred(), nil, nil); err != nil {
		// Unhealthy
		return nil, err
	}

	// For Dragonfly, no error returned means healthy
	return &DriverStatus{
		Status: DriverStatusHealthy,
	}, nil
}

// Preheat implements @Driver.Preheat.
func (dd *DragonflyDriver) Preheat(preheatingImage *PreheatImage) (*PreheatingStatus, error) {
	if dd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	if preheatingImage == nil {
		return nil, errors.New("no image specified")
	}

	var extraAttrs dragonflyExtraAttrs
	if preheatingImage.ExtraAttrs != nil && len(preheatingImage.ExtraAttrs) > 0 {
		extraAttrsStr, err := json.Marshal(preheatingImage.ExtraAttrs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal extra attributes")
		}

		if err := json.Unmarshal(extraAttrsStr, &extraAttrs); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal extra attributes")
		}
	}

	// Construct the preheat job request by the given parameters of the preheating image .
	req := &dragonflyCreateJobRequest{
		Type: "preheat",
		// TODO: Support set SchedulerClusterIDs, FilteredQueryParam, ConcurrentCount and Timeout.
		Args: dragonflyCreateJobRequestArgs{
			Type:    preheatingImage.Type,
			URL:     preheatingImage.URL,
			Headers: headerToMapString(preheatingImage.Headers),
			Scope:   preheatingImage.Scope,
		},
	}

	// Set the cluster ids if it is specified.
	if len(extraAttrs.ClusterIDs) > 0 {
		req.SchedulerClusterIDs = extraAttrs.ClusterIDs
	}

	url := fmt.Sprintf("%s%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), dragonflyJobPath)
	data, err := client.GetHTTPClient(dd.instance.Insecure).Post(url, dd.getCred(), req, nil)
	if err != nil {
		return nil, err
	}

	resp := &dragonflyJobResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	return &PreheatingStatus{
		TaskID:     fmt.Sprintf("%d", resp.ID),
		Status:     provider.PreheatingStatusPending,
		StartTime:  resp.CreatedAt.Format(time.RFC3339),
		FinishTime: resp.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// CheckProgress implements @Driver.CheckProgress.
func (dd *DragonflyDriver) CheckProgress(taskID string) (*PreheatingStatus, error) {
	if dd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	if taskID == "" {
		return nil, errors.New("no task ID")
	}

	url := fmt.Sprintf("%s%s/%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), dragonflyJobPath, taskID)
	data, err := client.GetHTTPClient(dd.instance.Insecure).Get(url, dd.getCred(), nil, nil)
	if err != nil {
		return nil, err
	}

	resp := &dragonflyJobResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	var (
		successMessage string
		errorMessage   string
	)

	var state string
	switch resp.State {
	case dragonflyJobPendingState:
		state = provider.PreheatingStatusRunning
	case dragonflyJobSuccessState:
		state = provider.PreheatingStatusSuccess

		var buffer bytes.Buffer
		table := tablewriter.NewWriter(&buffer)
		table.SetHeader([]string{"Blob URL", "Hostname", "IP", "Cluster ID", "State", "Error Message"})
		for _, jobState := range resp.Result.JobStates {
			for _, result := range jobState.Results {
				// Write the success tasks records to the table.
				for _, successTask := range result.SuccessTasks {
					table.Append([]string{successTask.URL, successTask.Hostname, successTask.IP, fmt.Sprint(result.SchedulerClusterID), dragonflyJobSuccessState, ""})
				}

				// Write the failure tasks records to the table.
				for _, failureTask := range result.FailureTasks {
					table.Append([]string{failureTask.URL, failureTask.Hostname, failureTask.IP, fmt.Sprint(result.SchedulerClusterID), dragonflyJobFailureState, failureTask.Description})
				}
			}
		}

		table.Render()
		successMessage = buffer.String()
	case dragonflyJobFailureState:
		var errs errors.Errors
		state = provider.PreheatingStatusFail
		for _, jobState := range resp.Result.JobStates {
			errs = append(errs, errors.New(jobState.Error))
		}

		if len(errs) > 0 {
			errorMessage = errs.Error()
		}
	default:
		state = provider.PreheatingStatusFail
		errorMessage = fmt.Sprintf("unknown state: %s", resp.State)
	}

	return &PreheatingStatus{
		TaskID:     fmt.Sprintf("%d", resp.ID),
		Status:     state,
		Message:    successMessage,
		Error:      errorMessage,
		StartTime:  resp.CreatedAt.Format(time.RFC3339),
		FinishTime: resp.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (dd *DragonflyDriver) getCred() *auth.Credential {
	return &auth.Credential{
		Mode: dd.instance.AuthMode,
		Data: dd.instance.AuthInfo,
	}
}

func headerToMapString(header map[string]interface{}) map[string]string {
	m := make(map[string]string)
	for k, v := range header {
		if s, ok := v.(string); ok {
			m[k] = s
		}
	}

	return m
}
