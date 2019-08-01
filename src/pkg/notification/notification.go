package notification

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification/config"
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
	SupportedEventTypes map[string]int

	// SupportedNotifyTypes is a map to store notification type, eg. HTTP, Email etc
	SupportedNotifyTypes map[string]int
)

// Init ...
func Init() {
	config.Config = &config.Configuration{
		CoreURL:          cfg.InternalCoreURL(),
		TokenServiceURL:  cfg.InternalTokenServiceEndpoint(),
		JobserviceURL:    cfg.InternalJobServiceURL(),
		CoreSecret:       cfg.CoreSecret(),
		JobserviceSecret: cfg.JobserviceSecret(),
	}

	// init notification policy manager
	PolicyMgr = manager.NewDefaultManger()
	// init hook manager
	HookManager = hook.NewHookManager()
	// init notification job manager
	JobMgr = jobMgr.NewDefaultManager()

	SupportedEventTypes = make(map[string]int)
	SupportedNotifyTypes = make(map[string]int)

	initSupportedEventType(
		model.EventTypePushImage, model.EventTypePullImage, model.EventTypeDeleteImage,
		model.EventTypeUploadChart, model.EventTypeDeleteChart, model.EventTypeDownloadChart,
		model.EventTypeScanningCompleted, model.EventTypeScanningFailed,
	)

	initSupportedNotifyType(model.NotifyTypeHTTP)

	log.Info("notification initialization completed")
}

func initSupportedEventType(eventTypes ...string) {
	for _, eventType := range eventTypes {
		SupportedEventTypes[eventType] = model.ValidType
	}
}

func initSupportedNotifyType(notifyTypes ...string) {
	for _, notifyType := range notifyTypes {
		SupportedNotifyTypes[notifyType] = model.ValidType
	}
}
