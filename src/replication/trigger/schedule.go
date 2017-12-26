package trigger

import (
	"fmt"
	"time"

	"github.com/vmware/harbor/src/common/scheduler"
	"github.com/vmware/harbor/src/common/scheduler/policy"
	replication_task "github.com/vmware/harbor/src/common/scheduler/task/replication"
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
	config := &policy.AlternatePolicyConfiguration{}
	switch st.params.Type {
	case replication.TriggerScheduleDaily:
		config.Duration = 24 * 3600 * time.Second
		config.OffsetTime = st.params.Offtime
	case replication.TriggerScheduleWeekly:
		config.Duration = 7 * 24 * 3600 * time.Second
		config.OffsetTime = st.params.Offtime
		config.Weekday = st.params.Weekday
	default:
		return fmt.Errorf("unsupported schedual trigger type: %s", st.params.Type)
	}

	schedulePolicy := policy.NewAlternatePolicy(assembleName(st.params.PolicyID), config)
	attachTask := replication_task.NewTask(st.params.PolicyID)
	schedulePolicy.AttachTasks(attachTask)
	return scheduler.DefaultScheduler.Schedule(schedulePolicy)
}

//Unset is the implementation of same method defined in Trigger interface
func (st *ScheduleTrigger) Unset() error {
	return scheduler.DefaultScheduler.UnSchedule(assembleName(st.params.PolicyID))
}

func assembleName(policyID int64) string {
	return fmt.Sprintf("replication_policy_%d", policyID)
}
