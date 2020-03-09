package topic

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/handler/auditlog"
	"github.com/goharbor/harbor/src/pkg/notifier/handler/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// Subscribe topics
func init() {
	handlersMap := map[string][]notifier.NotificationHandler{
		model.PushTagTopic:   {&auditlog.AuditHandler},
		model.PullTagTopic:   {&auditlog.AuditHandler},
		model.DeleteTagTopic: {&auditlog.AuditHandler},

		model.CreateProjectTopic: {&auditlog.AuditHandler},
		model.DeleteProjectTopic: {&auditlog.AuditHandler},

		model.CreateRepositoryTopic: {&auditlog.AuditHandler},
		model.DeleteRepositoryTopic: {&auditlog.AuditHandler},

		model.CreateArtifactTopic: {&auditlog.AuditHandler},
		model.DeleteArtifactTopic: {&auditlog.AuditHandler},

		model.WebhookTopic:           {&notification.HTTPHandler{}},
		model.UploadChartTopic:       {&notification.ChartPreprocessHandler{}},
		model.DownloadChartTopic:     {&notification.ChartPreprocessHandler{}},
		model.DeleteChartTopic:       {&notification.ChartPreprocessHandler{}},
		model.ScanningCompletedTopic: {&notification.ScanImagePreprocessHandler{}},
		model.ScanningFailedTopic:    {&notification.ScanImagePreprocessHandler{}},
		model.QuotaExceedTopic:       {&notification.QuotaPreprocessHandler{}},
	}

	for t, handlers := range handlersMap {
		for _, handler := range handlers {
			if err := notifier.Subscribe(t, handler); err != nil {
				log.Errorf("failed to subscribe topic %s: %v", t, err)
				continue
			}
			log.Debugf("topic %s is subscribed", t)
		}
	}
}
