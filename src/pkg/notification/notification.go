package notification

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/notification/hook"
	"github.com/goharbor/harbor/src/pkg/notification/job"
	jobMgr "github.com/goharbor/harbor/src/pkg/notification/job/manager"
	"github.com/goharbor/harbor/src/pkg/notification/model"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	"github.com/goharbor/harbor/src/pkg/notification/policy/manager"
)

var (
	// PolicyMgr is a global notification policy manager
	PolicyMgr policy.Manager

	// JobMgr is a notification job controller
	JobMgr job.Manager

	// HookManager is a hook manager
	HookManager hook.Manager

	// SupportedEventTypes is a map to store supported event type, eg. pushImage, pullImage etc
	SupportedEventTypes map[string]struct{}

	// SupportedNotifyTypes is a map to store notification type, eg. HTTP, Email etc
	SupportedNotifyTypes map[string]struct{}
)

// Init ...
func Init() {
	// init notification policy manager
	PolicyMgr = manager.NewDefaultManger()
	// init hook manager
	HookManager = hook.NewHookManager()
	// init notification job manager
	JobMgr = jobMgr.NewDefaultManager()

	SupportedEventTypes = make(map[string]struct{})
	SupportedNotifyTypes = make(map[string]struct{})

	initSupportedEventType(
		model.EventTypePushImage, model.EventTypePullImage, model.EventTypeDeleteImage,
		model.EventTypeUploadChart, model.EventTypeDeleteChart, model.EventTypeDownloadChart,
		model.EventTypeScanningCompleted, model.EventTypeScanningFailed, model.EventTypeProjectQuota,
	)

	initSupportedNotifyType(model.NotifyTypeHTTP)

	log.Info("notification initialization completed")
}

func initSupportedEventType(eventTypes ...string) {
	for _, eventType := range eventTypes {
		SupportedEventTypes[eventType] = struct{}{}
	}
}

func initSupportedNotifyType(notifyTypes ...string) {
	for _, notifyType := range notifyTypes {
		SupportedNotifyTypes[notifyType] = struct{}{}
	}
}
