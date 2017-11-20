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
		replication.TriggerKindImmediate,
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
			if err := ctl.triggerManager.SetupTrigger(policy.ID, *policy.Trigger); err != nil {
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
	//Validate policy
	// TODO

	return ctl.policyManager.CreatePolicy(newPolicy)
}

//UpdatePolicy will update the policy with new content.
//Parameter updatedPolicy must have the ID of the updated policy.
func (ctl *Controller) UpdatePolicy(updatedPolicy models.ReplicationPolicy) error {
	// TODO check pre-conditions
	return ctl.policyManager.UpdatePolicy(updatedPolicy)
}

//RemovePolicy will remove the specified policy and clean the related settings
func (ctl *Controller) RemovePolicy(policyID int64) error {
	// TODO check pre-conditions
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
func (ctl *Controller) Replicate(policyID int64) error {
	return nil
}
