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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/docker/distribution/manifest/schema2"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/notification"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/client"
)

const (
	krakenHealthPath  = "/health"
	krakenPreheatPath = "/registry/notifications"
)

// KrakenDriver implements the provider driver interface for Uber kraken.
// More details, please refer to https://github.com/uber/kraken
type KrakenDriver struct {
	instance *provider.Instance
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
	}
}

// GetHealth implements @Driver.GetHealth.
func (kd *KrakenDriver) GetHealth() (*DriverStatus, error) {
	if kd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	url := fmt.Sprintf("%s%s", strings.TrimSuffix(kd.instance.Endpoint, "/"), krakenHealthPath)
	url, err := lib.ValidateHTTPURL(url)
	if err != nil {
		return nil, err
	}
	_, err = client.GetHTTPClient(kd.instance.Insecure).Get(url, kd.getCred(), nil, nil)
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
	var events = make([]notification.Event, 0)
	eventID := utils.GenerateRandomString()
	event := notification.Event{
		ID:        eventID,
		TimeStamp: time.Now().UTC(),
		Action:    "push",
		Target: &notification.Target{
			MediaType:  schema2.MediaTypeManifest,
			Digest:     preheatingImage.Digest,
			Repository: preheatingImage.ImageName,
			URL:        preheatingImage.URL,
			Tag:        preheatingImage.Tag,
		},
	}
	events = append(events, event)
	var payload = notification.Notification{
		Events: events,
	}
	_, err := client.GetHTTPClient(kd.instance.Insecure).Post(url, kd.getCred(), payload, nil)
	if err != nil {
		return nil, err
	}

	return &PreheatingStatus{
		TaskID:     eventID,
		Status:     provider.PreheatingStatusSuccess,
		FinishTime: time.Now().String(),
	}, nil
}

// CheckProgress implements @Driver.CheckProgress.
// TODO: This should be improved later
func (kd *KrakenDriver) CheckProgress(taskID string) (*PreheatingStatus, error) {
	return &PreheatingStatus{
		TaskID:     taskID,
		Status:     provider.PreheatingStatusSuccess,
		FinishTime: time.Now().String(),
	}, nil
}

func (kd *KrakenDriver) getCred() *auth.Credential {
	return &auth.Credential{
		Mode: kd.instance.AuthMode,
		Data: kd.instance.AuthInfo,
	}
}
