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

	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/controller/replication"
	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
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
	var matches []*policyMatch
	var err error
	switch event.Type {
	case EventTypeArtifactPush, EventTypeTagDelete, EventTypeArtifactDelete:
		matches, err = getRelatedPolicies(ctx, event.Resource)
	default:
		return fmt.Errorf("unsupported event type %s", event.Type)
	}
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		log.Debugf("no policy found for the event %v, do nothing", event)
		return nil
	}

	if event.Operator != "" {
		ctx = context.WithValue(ctx, operator.ContextKey{}, event.Operator)
	}

	for _, match := range matches {
		id, err := replication.Ctl.Start(ctx, match.policy, match.resource, task.ExecutionTriggerEvent)
		if err != nil {
			return err
		}
		log.Debugf("%s event received, the replication execution %d started", event.Type, id)
	}
	return nil
}

type policyMatch struct {
	policy   *repctlmodel.Policy
	resource *model.Resource
}

func getRelatedPolicies(ctx context.Context, resource *model.Resource) ([]*policyMatch, error) {
	// Only query enabled event-based replication policies here, so the loop below
	// doesn't need to check policy.Enabled or policy.Trigger.Type again.
	policies, err := replication.Ctl.ListPolicies(ctx, q.New(q.KeyWords{
		"Enabled":            true,
		"Trigger__icontains": fmt.Sprintf("\"type\":\"%s\"", model.TriggerTypeEventBased),
	}))
	if err != nil {
		return nil, err
	}
	result := []*policyMatch{}
	for _, policy := range policies {
		// currently, the events are produced only by local Harbor,
		// so they should only apply to the policies whose source registry is local Harbor
		if !(policy.SrcRegistry == nil || policy.SrcRegistry.ID == 0) {
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

		result = append(result, &policyMatch{policy: policy, resource: resources[0]})
	}
	return result, nil
}
