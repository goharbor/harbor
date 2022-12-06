package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/client"
)

const (
	healthCheckEndpoint = "/_ping"
	preheatEndpoint     = "/preheats"
	preheatTaskEndpoint = "/preheats/{task_id}"
	dragonflyPending    = "WAITING"
	dragonflyFailed     = "FAILED"
)

type dragonflyPreheatCreateResp struct {
	ID string `json:"ID"`
}

type dragonflyPreheatInfo struct {
	ID         string `json:"ID"`
	StartTime  string `json:"startTime,omitempty"`
	FinishTime string `json:"finishTime,omitempty"`
	ErrorMsg   string `json:"errorMsg"`
	Status     string
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
		Icon:        "https://raw.githubusercontent.com/alibaba/Dragonfly/master/docs/images/logo.png",
		Version:     "0.10.1",
		Source:      "https://github.com/alibaba/Dragonfly",
		Maintainers: []string{"Jin Zhang/taiyun.zj@alibaba-inc.com"},
	}
}

// GetHealth implements @Driver.GetHealth.
func (dd *DragonflyDriver) GetHealth() (*DriverStatus, error) {
	if dd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	url := fmt.Sprintf("%s%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), healthCheckEndpoint)
	url, err := lib.ValidateHTTPURL(url)
	if err != nil {
		return nil, err
	}
	_, err = client.GetHTTPClient(dd.instance.Insecure).Get(url, dd.getCred(), nil, nil)
	if err != nil {
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

	taskStatus := provider.PreheatingStatusPending // default
	url := fmt.Sprintf("%s%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), preheatEndpoint)
	bytes, err := client.GetHTTPClient(dd.instance.Insecure).Post(url, dd.getCred(), preheatingImage, nil)
	if err != nil {
		if httpErr, ok := err.(*common_http.Error); ok && httpErr.Code == http.StatusAlreadyReported {
			// If the resource was preheated already with empty task ID, we should set preheat status to success.
			// Otherwise later querying for the task
			taskStatus = provider.PreheatingStatusSuccess
		} else {
			return nil, err
		}
	}

	result := &dragonflyPreheatCreateResp{}
	if err := json.Unmarshal(bytes, result); err != nil {
		return nil, err
	}

	return &PreheatingStatus{
		TaskID: result.ID,
		Status: taskStatus,
	}, nil
}

// CheckProgress implements @Driver.CheckProgress.
func (dd *DragonflyDriver) CheckProgress(taskID string) (*PreheatingStatus, error) {
	status, err := dd.getProgressStatus(taskID)
	if err != nil {
		return nil, err
	}

	// If preheat job already exists
	if strings.Index(status.ErrorMsg, "preheat task already exists, id:") >= 0 {
		if taskID, err = getTaskExistedFromErrMsg(status.ErrorMsg); err != nil {
			return nil, err
		}
		if status, err = dd.getProgressStatus(taskID); err != nil {
			return nil, err
		}
	}

	if status.Status == dragonflyPending {
		status.Status = provider.PreheatingStatusPending
	} else if status.Status == dragonflyFailed {
		status.Status = provider.PreheatingStatusFail
	}

	res := &PreheatingStatus{
		Status: status.Status,
		TaskID: taskID,
	}
	if status.StartTime != "" {
		res.StartTime = status.StartTime
	}
	if status.FinishTime != "" {
		res.FinishTime = status.FinishTime
	}

	return res, nil
}

func getTaskExistedFromErrMsg(msg string) (string, error) {
	begin := strings.Index(msg, "preheat task already exists, id:") + 32
	end := strings.LastIndex(msg, "\"}")
	if end-begin <= 0 {
		return "", errors.Errorf("can't find existed task id by error msg:%s", msg)
	}
	return msg[begin:end], nil
}

func (dd *DragonflyDriver) getProgressStatus(taskID string) (*dragonflyPreheatInfo, error) {
	if dd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	if len(taskID) == 0 {
		return nil, errors.New("no task ID")
	}

	path := strings.Replace(preheatTaskEndpoint, "{task_id}", taskID, 1)
	url := fmt.Sprintf("%s%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), path)
	bytes, err := client.GetHTTPClient(dd.instance.Insecure).Get(url, dd.getCred(), nil, nil)
	if err != nil {
		return nil, err
	}

	status := &dragonflyPreheatInfo{}
	if err := json.Unmarshal(bytes, status); err != nil {
		return nil, err
	}
	return status, nil
}

func (dd *DragonflyDriver) getCred() *auth.Credential {
	return &auth.Credential{
		Mode: dd.instance.AuthMode,
		Data: dd.instance.AuthInfo,
	}
}
