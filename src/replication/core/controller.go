package core

import (
	"fmt"

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
func (ctl *Controller) Replicate(policyID int64, metadate ...map[string]interface{}) error {

	fmt.Printf("replicating %d, metadata: %v ...\n", policyID, metadate)

	return nil
}
