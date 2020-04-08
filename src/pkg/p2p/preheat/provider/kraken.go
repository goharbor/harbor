package provider

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/docker/distribution/manifest/schema2"
	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/client"
	"github.com/goharbor/harbor/src/pkg/registry"
)

const (
	krakenHealthPath  = "/health"
	krakenPreheatPath = "/registry/notifications"
)

// KrakenDriver implements the provider driver interface for Uber kraken.
// More details, please refer to https://github.com/uber/kraken
type KrakenDriver struct {
	instance *models.Metadata
}

// Self implements @Driver.Self.
func (kd *KrakenDriver) Self() *Metadata {
	return &Metadata{
		ID:          "kraken",
		Name:        "Kraken",
		Icon:        "https://github.com/uber/kraken/blob/master/assets/kraken-logo-color.svg",
		Version:     "0.1.3",
		Source:      "https://github.com/uber/kraken",
		Maintainers: []string{"mmpei/peimingming@corp.netease.com"},
		AuthMode:    auth.AuthModeNone,
	}
}

// GetHealth implements @Driver.GetHealth.
func (kd *KrakenDriver) GetHealth() (*DriverStatus, error) {
	if kd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	url := fmt.Sprintf("%s%s", strings.TrimSuffix(kd.instance.Endpoint, "/"), krakenHealthPath)
	_, err := client.DefaultHTTPClient.Get(url, kd.getCred(), nil, nil)
	if err != nil {
		// Unhealthy
		return nil, err
	}

	// For Kraken, no error returned means healthy
	return &DriverStatus{
		Status: DriverStatusHealthy,
	}, nil
}

// Preheat implements @Driver.Preheat.
func (kd *KrakenDriver) Preheat(preheatingImage *PreheatImage) (*PreheatingStatus, error) {
	if kd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	if preheatingImage == nil {
		return nil, errors.New("no image specified")
	}

	url := fmt.Sprintf("%s%s", strings.TrimSuffix(kd.instance.Endpoint, "/"), krakenPreheatPath)
	var events = []common_models.Event{}
	eventID := utils.GenerateRandomString()
	digest, err := fetchDigest(preheatingImage.ImageName, preheatingImage.Tag)
	if err != nil {
		return nil, err
	}
	event := common_models.Event{
		ID:        eventID,
		TimeStamp: time.Now().UTC(),
		Action:    "push",
		Target: &common_models.Target{
			MediaType:  schema2.MediaTypeManifest,
			Digest:     digest,
			Repository: preheatingImage.ImageName,
			URL:        preheatingImage.URL,
			Tag:        preheatingImage.Tag,
		},
	}
	events = append(events, event)
	var payload = common_models.Notification{
		Events: events,
	}
	_, err = client.DefaultHTTPClient.Post(url, kd.getCred(), payload, nil)
	if err != nil {
		return nil, err
	}

	return &PreheatingStatus{
		TaskID:     eventID,
		Status:     models.PreheatingStatusSuccess,
		StartTime:  time.Now().String(),
		FinishTime: time.Now().String(),
	}, nil
}

// CheckProgress implements @Driver.CheckProgress.
func (kd *KrakenDriver) CheckProgress(taskID string) (*PreheatingStatus, error) {
	hr, err := dao.GetHistoryRecordByTaskID(taskID)
	if err != nil {
		return nil, err
	}

	if hr == nil {
		return nil, errors.New("preheat record not found")
	}

	return &PreheatingStatus{
		TaskID:     taskID,
		Status:     hr.Status,
		StartTime:  time.Now().String(),
		FinishTime: time.Now().String(),
	}, nil
}

func (kd *KrakenDriver) getCred() *auth.Credential {
	return &auth.Credential{
		Mode: kd.instance.AuthMode,
		Data: kd.instance.AuthData,
	}
}

func fetchDigest(repoName, tag string) (string, error) {
	exist, digest, err := registry.Cli.ManifestExist(repoName, tag)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", errors.New("image not found")
	}
	return digest, nil
}
