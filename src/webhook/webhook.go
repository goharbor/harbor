package webhook

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/webhook/config"
	"github.com/goharbor/harbor/src/webhook/hook"
	"github.com/goharbor/harbor/src/webhook/job"
	"github.com/goharbor/harbor/src/webhook/job/controller"
	"github.com/goharbor/harbor/src/webhook/model"
	"github.com/goharbor/harbor/src/webhook/policy"
)

var (
	// PolicyCtl is a global webhook policy controller
	PolicyCtl policy.Controller

	// JobCtl is a webhook job controller
	JobCtl job.Controller

	// HookManager is a hook manager
	HookManager hook.Manager

	// SupportedHookTypes is a map store support webhook type, eg. pushImage, pullImage etc
	SupportedHookTypes map[string]int

	// SupportedSendTypes is a map store webhook send type, eg. HTTP, Email etc
	SupportedSendTypes map[string]int
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

	// init webhook policy controller
	PolicyCtl = policy.NewController()
	// init hook manager
	HookManager = hook.NewHookManager()
	// init webhook job controller
	JobCtl = controller.NewController()

	SupportedHookTypes = make(map[string]int)
	SupportedSendTypes = make(map[string]int)

	initSupportedWebhookType(
		model.EventTypePushImage, model.EventTypePullImage, model.EventTypeDeleteImage,
		model.EventTypeUploadChart, model.EventTypeDeleteChart, model.EventTypeDownloadChart,
		model.EventTypeScanningCompleted, model.EventTypeScanningFailed,
	)

	initSupportedSendType(model.HookSendTypeHTTP)

	log.Info("webhook initialization completed")
}

func initSupportedWebhookType(hookTypes ...string) {
	for _, hookType := range hookTypes {
		SupportedHookTypes[hookType] = model.ValidType
	}
}

func initSupportedSendType(sendTypes ...string) {
	for _, sendType := range sendTypes {
		SupportedSendTypes[sendType] = model.ValidType
	}
}
