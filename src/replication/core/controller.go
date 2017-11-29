package core

import (
	"fmt"

	"github.com/vmware/harbor/src/common/dao"
	common_models "github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/models"
	"github.com/vmware/harbor/src/replication/policy"
	"github.com/vmware/harbor/src/replication/source"
	"github.com/vmware/harbor/src/replication/trigger"
)

//Controller is core module to cordinate and control the overall workflow of the
//replication modules.
type Controller struct {
	//Indicate whether the controller has been initialized or not
	initialized bool

	//Manage the policies
	policyManager *policy.Manager

	//Handle the things related with source
	sourcer *source.Sourcer

	//Manage the triggers of policies
	triggerManager *trigger.Manager
}

//Keep controller as singleton instance
var (
	DefaultController = NewController(ControllerConfig{}) //Use default data
)

//ControllerConfig includes related configurations required by the controller
type ControllerConfig struct {
	//The capacity of the cache storing enabled triggers
	CacheCapacity int
}

//NewController is the constructor of Controller.
func NewController(config ControllerConfig) *Controller {
	//Controller refer the default instances
	return &Controller{
		policyManager:  policy.NewManager(),
		sourcer:        source.NewSourcer(),
		triggerManager: trigger.NewManager(config.CacheCapacity),
	}
}

//Init will initialize the controller and the sub components
func (ctl *Controller) Init() error {
	if ctl.initialized {
		return nil
	}

	//Build query parameters
	triggerNames := []string{
		replication.TriggerKindSchedule,
	}
	queryName := ""
	for _, name := range triggerNames {
		queryName = fmt.Sprintf("%s,%s", queryName, name)
	}
	//Enable the triggers
	query := models.QueryParameter{
		TriggerName: queryName,
	}

	policies, err := ctl.policyManager.GetPolicies(query)
	if err != nil {
		return err
	}
	if policies != nil && len(policies) > 0 {
		for _, policy := range policies {
			if err := ctl.triggerManager.SetupTrigger(&policy); err != nil {
				//TODO: Log error
				fmt.Printf("Error: %s", err)
				//TODO:Update the status of policy
			}
		}
	}

	//Initialize sourcer
	ctl.sourcer.Init()

	ctl.initialized = true

	return nil
}

//CreatePolicy is used to create a new policy and enable it if necessary
func (ctl *Controller) CreatePolicy(newPolicy models.ReplicationPolicy) (int64, error) {
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

//UpdatePolicy will update the policy with new content.
//Parameter updatedPolicy must have the ID of the updated policy.
func (ctl *Controller) UpdatePolicy(updatedPolicy models.ReplicationPolicy) error {
	// TODO check pre-conditions

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
			if updatedPolicy.Trigger.Param != originPolicy.Trigger.Param {
				reset = true
			}
		case replication.TriggerKindImmediate:
			// Always reset immediate trigger as it is relevent with namespaces
			reset = true
		default:
			// manual trigger, no need to reset
		}
	}

	if err = ctl.policyManager.UpdatePolicy(updatedPolicy); err != nil {
		return err
	}

	if reset {
		if err = ctl.triggerManager.UnsetTrigger(id, *originPolicy.Trigger); err != nil {
			return err
		}

		return ctl.triggerManager.SetupTrigger(&updatedPolicy)
	}

	return nil
}

//RemovePolicy will remove the specified policy and clean the related settings
func (ctl *Controller) RemovePolicy(policyID int64) error {
	// TODO check pre-conditions

	policy, err := ctl.policyManager.GetPolicy(policyID)
	if err != nil {
		return err
	}

	if policy.ID == 0 {
		return fmt.Errorf("policy %d not found", policyID)
	}

	if err = ctl.triggerManager.UnsetTrigger(policyID, *policy.Trigger); err != nil {
		return err
	}

	return ctl.policyManager.RemovePolicy(policyID)
}

//GetPolicy is delegation of GetPolicy of Policy.Manager
func (ctl *Controller) GetPolicy(policyID int64) (models.ReplicationPolicy, error) {
	return ctl.policyManager.GetPolicy(policyID)
}

//GetPolicies is delegation of GetPoliciemodels.ReplicationPolicy{}s of Policy.Manager
func (ctl *Controller) GetPolicies(query models.QueryParameter) ([]models.ReplicationPolicy, error) {
	return ctl.policyManager.GetPolicies(query)
}

//Replicate starts one replication defined in the specified policy;
//Can be launched by the API layer and related triggers.
func (ctl *Controller) Replicate(policyID int64, metadata ...map[string]interface{}) error {
	policy, err := ctl.GetPolicy(policyID)
	if err != nil {
		return err
	}
	if policy.ID == 0 {
		return fmt.Errorf("policy %d not found", policyID)
	}

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

	// prepare candidates for replication
	candidates = getCandidates(&policy, ctl.sourcer, candidates...)

	targets := []*common_models.RepTarget{}
	for _, targetID := range policy.TargetIDs {
		target, err := dao.GetRepTarget(targetID)
		if err != nil {
			return err
		}
		targets = append(targets, target)
	}

	// TODO merge tags whose repository is same into one struct

	// call job service to do the replication
	return replicate(candidates, targets)
}

func getCandidates(policy *models.ReplicationPolicy, sourcer *source.Sourcer, candidates ...models.FilterItem) []models.FilterItem {
	if len(candidates) == 0 {
		for _, namespace := range policy.Namespaces {
			candidates = append(candidates, models.FilterItem{
				Kind:      replication.FilterItemKindProject,
				Value:     namespace,
				Operation: replication.OperationPush,
			})
		}
	}

	filterChain := buildFilterChain(policy, sourcer)

	return filterChain.DoFilter(candidates)
}

func buildFilterChain(policy *models.ReplicationPolicy, sourcer *source.Sourcer) source.FilterChain {
	filters := []source.Filter{}

	patternMap := map[string]string{}
	for _, f := range policy.Filters {
		patternMap[f.Kind] = f.Pattern
	}

	// TODO convert wildcard to regex expression
	projectPattern, exist := patternMap[replication.FilterItemKindProject]
	if !exist {
		projectPattern = replication.PatternMatchAll
	}

	repositoryPattern, exist := patternMap[replication.FilterItemKindRepository]
	if !exist {
		repositoryPattern = replication.PatternMatchAll
	}
	repositoryPattern = fmt.Sprintf("%s/%s", projectPattern, repositoryPattern)

	tagPattern, exist := patternMap[replication.FilterItemKindProject]
	if !exist {
		tagPattern = replication.PatternMatchAll
	}
	tagPattern = fmt.Sprintf("%s:%s", repositoryPattern, tagPattern)

	if policy.Trigger.Kind == replication.TriggerKindImmediate {
		// build filter chain for immediate trigger policy
		filters = append(filters,
			source.NewPatternFilter(replication.FilterItemKindTag, tagPattern))
	} else {
		// build filter chain for manual and schedule trigger policy

		// append project filter
		filters = append(filters,
			source.NewPatternFilter(replication.FilterItemKindProject, projectPattern))
		// append repository filter
		filters = append(filters,
			source.NewPatternFilter(replication.FilterItemKindRepository,
				repositoryPattern, source.NewRepositoryConvertor(sourcer.GetAdaptor(replication.AdaptorKindHarbor))))
		// append tag filter
		filters = append(filters,
			source.NewPatternFilter(replication.FilterItemKindTag,
				tagPattern, source.NewTagConvertor(sourcer.GetAdaptor(replication.AdaptorKindHarbor))))
	}

	return source.NewDefaultFilterChain(filters)
}

func replicate(candidates []models.FilterItem, targets []*common_models.RepTarget) error {
	// TODO

	log.Infof("replicate candidates %v to targets %v", candidates, targets)

	return nil
}
