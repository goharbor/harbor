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

package event

import (
	"context"
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/controller/replication"
	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/task"
)

// Handle ...
func Handle(ctx context.Context, event *Event) error {
	if event == nil || event.Resource == nil ||
		event.Resource.Metadata == nil ||
		len(event.Resource.Metadata.Artifacts) == 0 {
		return errors.New("invalid event")
	}
	var policies []*repctlmodel.Policy
	var err error
	switch event.Type {
	case EventTypeArtifactPush, EventTypeChartUpload, EventTypeTagDelete,
		EventTypeArtifactDelete, EventTypeChartDelete:
		policies, err = getRelatedPolicies(ctx, event.Resource)
	default:
		return fmt.Errorf("unsupported event type %s", event.Type)
	}
	if err != nil {
		return err
	}

	if len(policies) == 0 {
		log.Debugf("no policy found for the event %v, do nothing", event)
		return nil
	}

	for _, policy := range policies {
		id, err := replication.Ctl.Start(ctx, policy, event.Resource, task.ExecutionTriggerEvent)
		if err != nil {
			return err
		}
		log.Debugf("%s event received, the replication execution %d started", event.Type, id)
	}
	return nil
}

func getRelatedPolicies(ctx context.Context, resource *model.Resource) ([]*repctlmodel.Policy, error) {
	policies, err := replication.Ctl.ListPolicies(ctx, nil)
	if err != nil {
		return nil, err
	}
	result := []*repctlmodel.Policy{}
	for _, policy := range policies {
		// disabled
		if !policy.Enabled {
			continue
		}
		// currently, the events are produced only by local Harbor,
		// so they should only apply to the policies whose source registry is local Harbor
		if !(policy.SrcRegistry == nil || policy.SrcRegistry.ID == 0) {
			continue
		}
		// has no trigger
		if policy.Trigger == nil {
			continue
		}
		// trigger type isn't event based
		if policy.Trigger.Type != model.TriggerTypeEventBased {
			continue
		}
		// doesn't replicate deletion
		if resource.Deleted && !policy.ReplicateDeletion {
			continue
		}

		resources, err := filter.DoFilterResources([]*model.Resource{resource}, policy.Filters)
		if err != nil {
			return nil, err
		}
		// doesn't match the filters
		if len(resources) == 0 {
			continue
		}

		result = append(result, policy)
	}
	return result, nil
}
