package trigger

import (
	"errors"

	"github.com/vmware/harbor/src/replication"
)

//ScheduleTrigger will schedule a alternate policy to provide 'daily' and 'weekly' trigger ways.
type ScheduleTrigger struct {
	params ScheduleParam
}

//NewScheduleTrigger is constructor of ScheduleTrigger
func NewScheduleTrigger(params ScheduleParam) *ScheduleTrigger {
	return &ScheduleTrigger{
		params: params,
	}
}

//Kind is the implementation of same method defined in Trigger interface
func (st *ScheduleTrigger) Kind() string {
	return replication.TriggerKindSchedule
}

//Setup is the implementation of same method defined in Trigger interface
func (st *ScheduleTrigger) Setup() error {
	return errors.New("Not implemented")
}

//Unset is the implementation of same method defined in Trigger interface
func (st *ScheduleTrigger) Unset() error {
	return errors.New("Not implemented")
}
