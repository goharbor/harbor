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

package artifact

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/handler/util"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/notification"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
)

// Handler preprocess artifact event data
type Handler struct {
}

// Name ...
func (a *Handler) Name() string {
	return "ArtifactWebhook"
}

// Handle preprocess artifact event data and then publish hook event
func (a *Handler) Handle(ctx context.Context, value any) error {
	if !config.NotificationEnable(ctx) {
		log.Debug("notification feature is not enabled")
		return nil
	}

	switch v := value.(type) {
	case *event.PushArtifactEvent:
		return a.handle(ctx, v.ArtifactEvent)
	case *event.PullArtifactEvent:
		return a.handle(ctx, v.ArtifactEvent)
	case *event.DeleteArtifactEvent:
		return a.handle(ctx, v.ArtifactEvent)
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	return nil
}

// IsStateful ...
func (a *Handler) IsStateful() bool {
	return false
}

func (a *Handler) handle(ctx context.Context, event *event.ArtifactEvent) error {
	prj, err := project.Ctl.Get(ctx, event.Artifact.ProjectID, project.Metadata(true))
	if err != nil {
		log.Errorf("failed to get project: %d, error: %v", event.Artifact.ProjectID, err)
		return err
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(ctx, prj.ProjectID, event.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", event.EventType, err)
		return err
	}

	log.Debugf("find %d policies for %s event", len(policies), event.EventType)

	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", event.EventType, event)
		return nil
	}

	payload, err := a.constructArtifactPayload(ctx, event, prj)
	if err != nil {
		return err
	}

	err = util.SendHookWithPolicies(ctx, policies, payload, event.EventType)
	if err != nil {
		return err
	}
	return nil
}

func (a *Handler) constructArtifactPayload(ctx context.Context, event *event.ArtifactEvent, project *proModels.Project) (*notifyModel.Payload, error) {
	repoName := event.Repository
	if repoName == "" {
		return nil, fmt.Errorf("invalid %s event with empty repo name", event.EventType)
	}

	repoType := proModels.ProjectPrivate
	if project.IsPublic() {
		repoType = proModels.ProjectPublic
	}

	imageName := util.GetNameFromImgRepoFullName(repoName)

	payload := &notifyModel.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &notifyModel.EventData{
			Repository: &notifyModel.Repository{
				Name:         imageName,
				Namespace:    project.Name,
				RepoFullName: repoName,
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	repoRecord, err := pkg.RepositoryMgr.GetByName(ctx, repoName)
	if err != nil {
		log.Errorf("failed to get repository with name %s: %v", repoName, err)
	} else {
		// for the delete repository event, it cannot get the repo info here, just let the creation time be empty.
		payload.EventData.Repository.DateCreated = repoRecord.CreationTime.Unix()
	}

	var reference string
	if len(event.Tags) == 0 {
		reference = event.Artifact.Digest
	} else {
		reference = event.Tags[0]
	}
	resURL, err := util.BuildImageResourceURL(repoName, reference)
	if err != nil {
		log.Errorf("get resource URL failed: %v", err)
		return nil, err
	}

	resource := &notifyModel.Resource{
		Tag:         reference,
		Digest:      event.Artifact.Digest,
		ResourceURL: resURL,
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)

	return payload, nil
}

// RobotHandler preprocess robot event data
type RobotHandler struct {
}

// Name ...
func (r *RobotHandler) Name() string {
	return "RobotWebhook"
}

// Handle preprocess robot event data and then publish hook event
func (r *RobotHandler) Handle(ctx context.Context, value any) error {
	if !config.NotificationEnable(ctx) {
		log.Debug("notification feature is not enabled")
		return nil
	}

	switch v := value.(type) {
	case *event.RobotExpiredEvent:
		return r.handleRobotExpired(ctx, v)
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	return nil
}

// IsStateful ...
func (r *RobotHandler) IsStateful() bool {
	return false
}

func (r *RobotHandler) handleRobotExpired(ctx context.Context, event *event.RobotExpiredEvent) error {
	// For robot expiration, we use project ID 0 for system-level webhooks
	// or the robot's project ID for project-level webhooks
	projectID := event.Robot.ProjectID
	if projectID == 0 {
		projectID = 0 // System-level robot
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(ctx, projectID, event.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", event.EventType, err)
		return err
	}

	log.Debugf("find %d policies for %s event", len(policies), event.EventType)

	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", event.EventType, event)
		return nil
	}

	payload, err := r.constructRobotPayload(ctx, event)
	if err != nil {
		return err
	}

	err = util.SendHookWithPolicies(ctx, policies, payload, event.EventType)
	if err != nil {
		return err
	}
	return nil
}

func (r *RobotHandler) constructRobotPayload(ctx context.Context, event *event.RobotExpiredEvent) (*notifyModel.Payload, error) {
	payload := &notifyModel.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &notifyModel.EventData{
			Custom: map[string]string{
				"robot_name":    event.Robot.Name,
				"robot_id":      fmt.Sprintf("%d", event.Robot.ID),
				"project_id":    fmt.Sprintf("%d", event.Robot.ProjectID),
				"expires_at":    fmt.Sprintf("%d", event.Robot.ExpiresAt),
				"creation_time": event.Robot.CreationTime.Format("2006-01-02 15:04:05"),
				"disabled_time": event.OccurAt.Format("2006-01-02 15:04:05"),
				"description":   event.Robot.Description,
			},
		},
		Operator: event.Operator,
	}

	return payload, nil
}
