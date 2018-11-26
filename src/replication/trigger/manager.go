package trigger

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/models"
)

// Manager provides unified methods to manage the triggers of policies;
// Cache the enabled triggers, setup/unset the trigger based on the parameters
// with json format.
type Manager struct {
	// Cache for triggers
	// cache *Cache
}

// NewManager is the constructor of trigger manager.
// capacity is the max number of trigger references manager can keep in memory
func NewManager(capacity int) *Manager {
	return &Manager{
		// cache: NewCache(capacity),
	}
}

/*
// GetTrigger returns the enabled trigger reference if existing in the cache.
func (m *Manager) GetTrigger(policyID int64) Interface {
	return m.cache.Get(policyID)
}

// RemoveTrigger will disable the trigger and remove it from the cache if existing.
func (m *Manager) RemoveTrigger(policyID int64) error {
	trigger := m.cache.Get(policyID)
	if trigger == nil {
		return errors.New("Trigger is not cached, please use UnsetTrigger to disable the trigger")
	}

	// Unset trigger
	if err := trigger.Unset(); err != nil {
		return err
	}

	// Remove from cache
	// No need to check the return of remove because the dirty item cached in the cache
	// will be removed out finally after a certain while
	m.cache.Remove(policyID)

	return nil
}
*/

// SetupTrigger will create the new trigger based on the provided policy.
// If failed, an error will be returned.
func (m *Manager) SetupTrigger(policy *models.ReplicationPolicy) error {
	trigger, err := createTrigger(policy)
	if err != nil {
		return err
	}

	// manual trigger, do nothing
	if trigger == nil {
		return nil
	}

	tg := trigger.(Interface)
	if err = tg.Setup(); err != nil {
		return err
	}

	log.Debugf("%s trigger for policy %d is set", tg.Kind(), policy.ID)
	return nil
}

// UnsetTrigger will disable the trigger which is not cached in the trigger cache.
func (m *Manager) UnsetTrigger(policy *models.ReplicationPolicy) error {
	trigger, err := createTrigger(policy)
	if err != nil {
		return err
	}

	// manual trigger, do nothing
	if trigger == nil {
		return nil
	}

	tg := trigger.(Interface)
	if err = tg.Unset(); err != nil {
		return err
	}

	log.Debugf("%s trigger for policy %d is unset", tg.Kind(), policy.ID)
	return nil
}

func createTrigger(policy *models.ReplicationPolicy) (interface{}, error) {
	if policy == nil || policy.Trigger == nil {
		return nil, fmt.Errorf("empty policy or trigger")
	}

	trigger := policy.Trigger
	switch trigger.Kind {
	case replication.TriggerKindSchedule:
		param := ScheduleParam{}
		param.PolicyID = policy.ID
		param.Type = trigger.ScheduleParam.Type
		param.Weekday = trigger.ScheduleParam.Weekday
		param.Offtime = trigger.ScheduleParam.Offtime

		return NewScheduleTrigger(param), nil
	case replication.TriggerKindImmediate:
		param := ImmediateParam{}
		param.PolicyID = policy.ID
		param.OnDeletion = policy.ReplicateDeletion
		param.Namespaces = policy.Namespaces

		return NewImmediateTrigger(param), nil
	case replication.TriggerKindManual:
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid trigger type: %s", trigger.Kind)
	}
}
