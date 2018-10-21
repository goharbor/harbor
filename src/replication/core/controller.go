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

package core

import (
	"fmt"
	"reflect"
	"strings"

	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/models"
	"github.com/goharbor/harbor/src/replication/policy"
	"github.com/goharbor/harbor/src/replication/replicator"
	"github.com/goharbor/harbor/src/replication/source"
	"github.com/goharbor/harbor/src/replication/target"
	"github.com/goharbor/harbor/src/replication/trigger"

	"github.com/docker/distribution/uuid"
)

// Controller defines the methods that a replicatoin controllter should implement
type Controller interface {
	policy.Manager
	Init() error
	Replicate(policyID int64, metadata ...map[string]interface{}) error
}

// DefaultController is core module to cordinate and control the overall workflow of the
// replication modules.
type DefaultController struct {
	// Indicate whether the controller has been initialized or not
	initialized bool

	// Manage the policies
	policyManager policy.Manager

	// Manage the targets
	targetManager target.Manager

	// Handle the things related with source
	sourcer *source.Sourcer

	// Manage the triggers of policies
	triggerManager *trigger.Manager

	// Handle the replication work
	replicator replicator.Replicator
}

// Keep controller as singleton instance
var (
	GlobalController Controller
)

// ControllerConfig includes related configurations required by the controller
type ControllerConfig struct {
	// The capacity of the cache storing enabled triggers
	CacheCapacity int
}

// NewDefaultController is the constructor of DefaultController.
func NewDefaultController(cfg ControllerConfig) *DefaultController {
	// Controller refer the default instances
	ctl := &DefaultController{
		policyManager:  policy.NewDefaultManager(),
		targetManager:  target.NewDefaultManager(),
		sourcer:        source.NewSourcer(),
		triggerManager: trigger.NewManager(cfg.CacheCapacity),
	}

	ctl.replicator = replicator.NewDefaultReplicator(utils.GetJobServiceClient())

	return ctl
}

// Init creates the GlobalController and inits it
func Init() error {
	GlobalController = NewDefaultController(ControllerConfig{}) // Use default data
	return GlobalController.Init()
}

// Init will initialize the controller and the sub components
func (ctl *DefaultController) Init() error {
	if ctl.initialized {
		return nil
	}

	// Initialize sourcer
	ctl.sourcer.Init()

	ctl.initialized = true

	return nil
}

// CreatePolicy is used to create a new policy and enable it if necessary
func (ctl *DefaultController) CreatePolicy(newPolicy models.ReplicationPolicy) (int64, error) {
	id, err := ctl.policyManager.CreatePolicy(newPolicy)
	if err != nil {
		return 0, err
	}

	newPolicy.ID = id
	if err = ctl.triggerManager.SetupTrigger(&newPolicy); err != nil {
		return 0, err
	}

	return id, nil
}

// UpdatePolicy will update the policy with new content.
// Parameter updatedPolicy must have the ID of the updated policy.
func (ctl *DefaultController) UpdatePolicy(updatedPolicy models.ReplicationPolicy) error {
	id := updatedPolicy.ID
	originPolicy, err := ctl.policyManager.GetPolicy(id)
	if err != nil {
		return err
	}

	if originPolicy.ID == 0 {
		return fmt.Errorf("policy %d not found", id)
	}

	reset := false
	if updatedPolicy.Trigger.Kind != originPolicy.Trigger.Kind {
		reset = true
	} else {
		switch updatedPolicy.Trigger.Kind {
		case replication.TriggerKindSchedule:
			if !originPolicy.Trigger.ScheduleParam.Equal(updatedPolicy.Trigger.ScheduleParam) {
				reset = true
			}
		case replication.TriggerKindImmediate:
			// Always reset immediate trigger as it is relevant with namespaces
			reset = true
		default:
			// manual trigger, no need to reset
		}
	}

	if err = ctl.policyManager.UpdatePolicy(updatedPolicy); err != nil {
		return err
	}

	if reset {
		if err = ctl.triggerManager.UnsetTrigger(&originPolicy); err != nil {
			return err
		}

		return ctl.triggerManager.SetupTrigger(&updatedPolicy)
	}

	return nil
}

// RemovePolicy will remove the specified policy and clean the related settings
func (ctl *DefaultController) RemovePolicy(policyID int64) error {
	// TODO check pre-conditions

	policy, err := ctl.policyManager.GetPolicy(policyID)
	if err != nil {
		return err
	}

	if policy.ID == 0 {
		return fmt.Errorf("policy %d not found", policyID)
	}

	if err = ctl.triggerManager.UnsetTrigger(&policy); err != nil {
		return err
	}

	return ctl.policyManager.RemovePolicy(policyID)
}

// GetPolicy is delegation of GetPolicy of Policy.Manager
func (ctl *DefaultController) GetPolicy(policyID int64) (models.ReplicationPolicy, error) {
	return ctl.policyManager.GetPolicy(policyID)
}

// GetPolicies is delegation of GetPoliciemodels.ReplicationPolicy{}s of Policy.Manager
func (ctl *DefaultController) GetPolicies(query models.QueryParameter) (*models.ReplicationPolicyQueryResult, error) {
	return ctl.policyManager.GetPolicies(query)
}

// Replicate starts one replication defined in the specified policy;
// Can be launched by the API layer and related triggers.
func (ctl *DefaultController) Replicate(policyID int64, metadata ...map[string]interface{}) error {
	policy, err := ctl.GetPolicy(policyID)
	if err != nil {
		return err
	}
	if policy.ID == 0 {
		return fmt.Errorf("policy %d not found", policyID)
	}

	// prepare candidates for replication
	candidates := getCandidates(&policy, ctl.sourcer, metadata...)
	if len(candidates) == 0 {
		log.Debugf("replication candidates are null, no further action needed")
	}

	targets := []*common_models.RepTarget{}
	for _, targetID := range policy.TargetIDs {
		target, err := ctl.targetManager.GetTarget(targetID)
		if err != nil {
			return err
		}
		targets = append(targets, target)
	}

	// Get operation uuid from metadata, if none provided, generate one.
	opUUID, err := getOpUUID(metadata...)
	if err != nil {
		return err
	}

	// submit the replication
	return ctl.replicator.Replicate(&replicator.Replication{
		PolicyID:   policyID,
		OpUUID:     opUUID,
		Candidates: candidates,
		Targets:    targets,
	})
}

func getCandidates(policy *models.ReplicationPolicy, sourcer *source.Sourcer,
	metadata ...map[string]interface{}) []models.FilterItem {
	candidates := []models.FilterItem{}
	if len(metadata) > 0 {
		meta := metadata[0]["candidates"]
		if meta != nil {
			cands, ok := meta.([]models.FilterItem)
			if ok {
				candidates = append(candidates, cands...)
			}
		}
	}

	if len(candidates) == 0 {
		for _, namespace := range policy.Namespaces {
			candidates = append(candidates, models.FilterItem{
				Kind:      replication.FilterItemKindProject,
				Value:     namespace,
				Operation: common_models.RepOpTransfer,
			})
		}
	}

	filterChain := buildFilterChain(policy, sourcer)

	return filterChain.DoFilter(candidates)
}

func buildFilterChain(policy *models.ReplicationPolicy, sourcer *source.Sourcer) source.FilterChain {
	filters := []source.Filter{}

	fm := map[string][]models.Filter{}
	for _, filter := range policy.Filters {
		fm[filter.Kind] = append(fm[filter.Kind], filter)
	}

	registry := sourcer.GetAdaptor(replication.AdaptorKindHarbor)
	// repository filter
	pattern := ""
	repoFilters := fm[replication.FilterItemKindRepository]
	if len(repoFilters) > 0 {
		pattern = repoFilters[0].Value.(string)
	}
	filters = append(filters,
		source.NewRepositoryFilter(pattern, registry))
	// tag filter
	pattern = ""
	tagFilters := fm[replication.FilterItemKindTag]
	if len(tagFilters) > 0 {
		pattern = tagFilters[0].Value.(string)
	}
	filters = append(filters,
		source.NewTagFilter(pattern, registry))
	// label filters
	var labelID int64
	for _, labelFilter := range fm[replication.FilterItemKindLabel] {
		labelID = labelFilter.Value.(int64)
		filters = append(filters, source.NewLabelFilter(labelID))
	}

	return source.NewDefaultFilterChain(filters)
}

// getOpUUID get operation uuid from metadata or generate one if none found.
func getOpUUID(metadata ...map[string]interface{}) (string, error) {
	if len(metadata) <= 0 {
		return strings.Replace(uuid.Generate().String(), "-", "", -1), nil
	}

	opUUID, ok := metadata[0]["op_uuid"]
	if !ok {
		return strings.Replace(uuid.Generate().String(), "-", "", -1), nil
	}

	id, ok := opUUID.(string)
	if !ok {
		return "", fmt.Errorf("operation uuid should have type 'string', but got '%s'", reflect.TypeOf(opUUID).Name())
	}

	if id == "" {
		return "", fmt.Errorf("provided operation uuid is empty")
	}

	return id, nil
}
