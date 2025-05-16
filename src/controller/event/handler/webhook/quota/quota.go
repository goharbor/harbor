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

package quota

import (
	"context"
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/handler/util"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
)

// Handler preprocess image event data
type Handler struct {
}

// Name ...
func (qp *Handler) Name() string {
	return "QuotaWebhook"
}

// Handle ...
func (qp *Handler) Handle(ctx context.Context, value any) error {
	quotaEvent, ok := value.(*event.QuotaEvent)
	if !ok {
		return errors.New("invalid quota event type")
	}
	if quotaEvent == nil {
		return fmt.Errorf("nil quota event")
	}

	prj, err := project.Ctl.GetByName(ctx, quotaEvent.Project.Name, project.Metadata(true))
	if err != nil {
		log.Errorf("failed to get project:%s, error: %v", quotaEvent.Project.Name, err)
		return err
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(ctx, prj.ProjectID, quotaEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", quotaEvent.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", quotaEvent.EventType, quotaEvent)
		return nil
	}

	payload, err := constructQuotaPayload(quotaEvent)
	if err != nil {
		return err
	}

	err = util.SendHookWithPolicies(ctx, policies, payload, quotaEvent.EventType)
	if err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (qp *Handler) IsStateful() bool {
	return false
}

func constructQuotaPayload(event *event.QuotaEvent) (*notifyModel.Payload, error) {
	repoName := event.RepoName
	if repoName == "" {
		return nil, fmt.Errorf("invalid %s event with empty repo name", event.EventType)
	}

	repoType := proModels.ProjectPrivate
	if event.Project.IsPublic() {
		repoType = proModels.ProjectPublic
	}

	imageName := util.GetNameFromImgRepoFullName(repoName)
	quotaCustom := make(map[string]string)
	quotaCustom["Details"] = event.Msg

	payload := &notifyModel.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &notifyModel.EventData{
			Repository: &notifyModel.Repository{
				Name:         imageName,
				Namespace:    event.Project.Name,
				RepoFullName: repoName,
				RepoType:     repoType,
			},
			Custom: quotaCustom,
		},
		Operator: event.Operator,
	}

	if event.Resource != nil {
		resource := &notifyModel.Resource{
			Tag:    event.Resource.Tag,
			Digest: event.Resource.Digest,
		}
		payload.EventData.Resources = append(payload.EventData.Resources, resource)
	}

	return payload, nil
}
