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

package util

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/distribution"
	policy_model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
)

// SendHookWithPolicies send hook by publishing topic of specified target type(notify type)
func SendHookWithPolicies(ctx context.Context, policies []*policy_model.Policy, payload *notifyModel.Payload, eventType string) error {
	// if global notification configured disabled, return directly
	if !config.NotificationEnable(ctx) {
		log.Debug("notification feature is not enabled")
		return nil
	}

	errRet := false
	for _, ply := range policies {
		targets := ply.Targets
		for _, target := range targets {
			evt := &event.Event{}
			hookMetadata := &event.HookMetaData{
				ProjectID: ply.ProjectID,
				EventType: eventType,
				PolicyID:  ply.ID,
				Payload:   payload,
				Target:    &target,
			}
			// It should never affect evaluating other policies when one is failed, but error should return
			if err := evt.Build(ctx, hookMetadata); err == nil {
				if err := evt.Publish(ctx); err != nil {
					errRet = true
					log.Errorf("failed to publish hook notify event: %v", err)
				}
			} else {
				errRet = true
				log.Errorf("failed to build hook notify event metadata: %v", err)
			}
			log.Debugf("published image event %s by topic %s", payload.Type, target.Type)
		}
	}
	if errRet {
		return errors.New("failed to trigger some of the events")
	}
	return nil
}

// GetNameFromImgRepoFullName gets image name from repo full name with format `repoName/imageName`
func GetNameFromImgRepoFullName(repo string) string {
	_, after, _ := strings.Cut(repo, "/")
	return after
}

// BuildImageResourceURL ...
func BuildImageResourceURL(repoName, reference string) (string, error) {
	extURL, err := config.ExtURL()
	if err != nil {
		return "", fmt.Errorf("get external endpoint failed: %v", err)
	}

	if distribution.IsDigest(reference) {
		return fmt.Sprintf("%s/%s@%s", extURL, repoName, reference), nil
	}

	return fmt.Sprintf("%s/%s:%s", extURL, repoName, reference), nil
}
