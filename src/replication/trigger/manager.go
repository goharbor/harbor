package trigger

import (
	"errors"

	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/models"
)

//Manager provides unified methods to manage the triggers of policies;
//Cache the enabled triggers, setup/unset the trigger based on the parameters
//with json format.
type Manager struct {
	//Cache for triggers
	//cache *Cache
}

//NewManager is the constructor of trigger manager.
//capacity is the max number of trigger references manager can keep in memory
func NewManager(capacity int) *Manager {
	return &Manager{
	//cache: NewCache(capacity),
	}
}

/*
//GetTrigger returns the enabled trigger reference if existing in the cache.
func (m *Manager) GetTrigger(policyID int64) Interface {
	return m.cache.Get(policyID)
}

//RemoveTrigger will disable the trigger and remove it from the cache if existing.
func (m *Manager) RemoveTrigger(policyID int64) error {
	trigger := m.cache.Get(policyID)
	if trigger == nil {
		return errors.New("Trigger is not cached, please use UnsetTrigger to disable the trigger")
	}

	//Unset trigger
	if err := trigger.Unset(); err != nil {
		return err
	}

	//Remove from cache
	//No need to check the return of remove because the dirty item cached in the cache
	//will be removed out finally after a certain while
	m.cache.Remove(policyID)

	return nil
}
*/

//SetupTrigger will create the new trigger based on the provided json parameters.
//If failed, an error will be returned.
func (m *Manager) SetupTrigger(policyID int64, trigger models.Trigger) error {
	if policyID <= 0 {
		return errors.New("Invalid policy ID")
	}

	if len(trigger.Kind) == 0 {
		return errors.New("Invalid replication trigger definition")
	}

	switch trigger.Kind {
	case replication.TriggerKindSchedule:
		param := ScheduleParam{}
		if err := param.Parse(trigger.Param); err != nil {
			return err
		}
		//Append policy ID info
		param.PolicyID = policyID

		newTrigger := NewScheduleTrigger(param)
		if err := newTrigger.Setup(); err != nil {
			return err
		}
	case replication.TriggerKindImmediate:
		param := ImmediateParam{}
		if err := param.Parse(trigger.Param); err != nil {
			return err
		}
		//Append policy ID info
		param.PolicyID = policyID

		newTrigger := NewImmediateTrigger(param)
		if err := newTrigger.Setup(); err != nil {
			return err
		}
	default:
		//Treat as manual trigger
		break
	}

	return nil
}

//UnsetTrigger will disable the trigger which is not cached in the trigger cache.
func (m *Manager) UnsetTrigger(policyID int64, trigger models.Trigger) error {
	if policyID <= 0 {
		return errors.New("Invalid policy ID")
	}

	if len(trigger.Kind) == 0 {
		return errors.New("Invalid replication trigger definition")
	}

	switch trigger.Kind {
	case replication.TriggerKindSchedule:
		param := ScheduleParam{}
		if err := param.Parse(trigger.Param); err != nil {
			return err
		}
		//Append policy ID info
		param.PolicyID = policyID

		newTrigger := NewScheduleTrigger(param)
		if err := newTrigger.Unset(); err != nil {
			return err
		}
	case replication.TriggerKindImmediate:
		param := ImmediateParam{}
		if err := param.Parse(trigger.Param); err != nil {
			return err
		}
		//Append policy ID info
		param.PolicyID = policyID

		newTrigger := NewImmediateTrigger(param)
		if err := newTrigger.Unset(); err != nil {
			return err
		}
	default:
		//Treat as manual trigger
		break
	}

	return nil
}
