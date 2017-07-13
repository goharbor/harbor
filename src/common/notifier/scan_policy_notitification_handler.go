package notifier

import (
	"errors"
	"reflect"

	"time"

	"github.com/vmware/harbor/src/common/scheduler"
	"github.com/vmware/harbor/src/common/scheduler/policy"
	"github.com/vmware/harbor/src/common/scheduler/task"
)

const (
	//PolicyTypeDaily specify the policy type is "daily"
	PolicyTypeDaily = "daily"

	alternatePolicy = "Alternate Policy"
)

//ScanPolicyNotification is defined for pass the policy change data.
type ScanPolicyNotification struct {
	//Type is used to keep the scan policy type: "none","daily" and "refresh".
	Type string

	//DailyTime is used when the type is 'daily', the offset with UTC time 00:00.
	DailyTime int64
}

//ScanPolicyNotificationHandler is defined to handle the changes of scanning
//policy.
type ScanPolicyNotificationHandler struct{}

//IsStateful to indicate this handler is stateful.
func (s *ScanPolicyNotificationHandler) IsStateful() bool {
	//Policy change should be done one by one.
	return true
}

//Handle the policy change notification.
func (s *ScanPolicyNotificationHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("ScanPolicyNotificationHandler can not handle nil value")
	}

	if reflect.TypeOf(value).Kind() != reflect.Struct ||
		reflect.TypeOf(value).String() != "notifier.ScanPolicyNotification" {
		return errors.New("ScanPolicyNotificationHandler can not handle value with invalid type")
	}

	notification := value.(ScanPolicyNotification)

	hasScheduled := scheduler.DefaultScheduler.HasScheduled(alternatePolicy)
	if notification.Type == PolicyTypeDaily {
		if !hasScheduled {
			schedulePolicy := policy.NewAlternatePolicy(&policy.AlternatePolicyConfiguration{
				Duration:   24 * time.Hour,
				OffsetTime: notification.DailyTime,
			})
			attachTask := task.NewScanAllTask()
			schedulePolicy.AttachTasks(attachTask)

			return scheduler.DefaultScheduler.Schedule(schedulePolicy)
		}
	} else {
		if hasScheduled {
			return scheduler.DefaultScheduler.UnSchedule(alternatePolicy)
		}
	}

	return nil
}
