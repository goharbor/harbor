package notifier

import (
	"errors"
	"reflect"

	"fmt"
	"time"

	"github.com/vmware/harbor/src/common/scheduler"
	"github.com/vmware/harbor/src/common/scheduler/policy"
	"github.com/vmware/harbor/src/common/scheduler/task"
)

const (
	//PolicyTypeDaily specify the policy type is "daily"
	PolicyTypeDaily = "daily"

	//PolicyTypeNone specify the policy type is "none"
	PolicyTypeNone = "none"

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
			//Schedule a new policy.
			return schedulePolicy(notification)
		}

		//To check and compare if the related parameter is changed.
		if pl := scheduler.DefaultScheduler.GetPolicy(alternatePolicy); pl != nil {
			policyCandidate := policy.NewAlternatePolicy(alternatePolicy, &policy.AlternatePolicyConfiguration{
				Duration:   24 * time.Hour,
				OffsetTime: notification.DailyTime,
			})
			if !pl.Equal(policyCandidate) {
				//Parameter changed.
				//Unschedule policy.
				if err := scheduler.DefaultScheduler.UnSchedule(alternatePolicy); err != nil {
					return err
				}

				//Schedule a new policy.
				return schedulePolicy(notification)
			}
			//Same policy configuration, do nothing
			return nil
		}

		return errors.New("Inconsistent policy scheduling status")
	} else if notification.Type == PolicyTypeNone {
		if hasScheduled {
			return scheduler.DefaultScheduler.UnSchedule(alternatePolicy)
		}
	} else {
		return fmt.Errorf("Notification type %s is not supported", notification.Type)
	}

	return nil
}

//Schedule policy.
func schedulePolicy(notification ScanPolicyNotification) error {
	schedulePolicy := policy.NewAlternatePolicy(alternatePolicy, &policy.AlternatePolicyConfiguration{
		Duration:   24 * time.Hour,
		OffsetTime: notification.DailyTime,
	})
	attachTask := task.NewScanAllTask()
	if err := schedulePolicy.AttachTasks(attachTask); err != nil {
		return err
	}

	return scheduler.DefaultScheduler.Schedule(schedulePolicy)
}
