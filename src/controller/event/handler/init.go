package handler

import (
	"context"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/handler/auditlog"
	"github.com/goharbor/harbor/src/controller/event/handler/internal"
	"github.com/goharbor/harbor/src/controller/event/handler/p2p"
	"github.com/goharbor/harbor/src/controller/event/handler/replication"
	"github.com/goharbor/harbor/src/controller/event/handler/webhook/artifact"
	"github.com/goharbor/harbor/src/controller/event/handler/webhook/quota"
	"github.com/goharbor/harbor/src/controller/event/handler/webhook/scan"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/task"
)

func init() {
	// notification
	_ = notifier.Subscribe(event.TopicPushArtifact, &artifact.Handler{})
	_ = notifier.Subscribe(event.TopicPullArtifact, &artifact.Handler{})
	_ = notifier.Subscribe(event.TopicDeleteArtifact, &artifact.Handler{})
	_ = notifier.Subscribe(event.TopicQuotaExceed, &quota.Handler{})
	_ = notifier.Subscribe(event.TopicQuotaWarning, &quota.Handler{})
	_ = notifier.Subscribe(event.TopicScanningFailed, &scan.Handler{})
	_ = notifier.Subscribe(event.TopicScanningStopped, &scan.Handler{})
	_ = notifier.Subscribe(event.TopicScanningCompleted, &scan.Handler{})
	_ = notifier.Subscribe(event.TopicDeleteArtifact, &scan.DelArtHandler{})
	_ = notifier.Subscribe(event.TopicReplication, &artifact.ReplicationHandler{})
	_ = notifier.Subscribe(event.TopicTagRetention, &artifact.RetentionHandler{})

	// replication
	_ = notifier.Subscribe(event.TopicPushArtifact, &replication.Handler{})
	_ = notifier.Subscribe(event.TopicDeleteArtifact, &replication.Handler{})
	_ = notifier.Subscribe(event.TopicCreateTag, &replication.Handler{})
	_ = notifier.Subscribe(event.TopicDeleteTag, &replication.Handler{})

	// p2p preheat
	_ = notifier.Subscribe(event.TopicPushArtifact, &p2p.Handler{})
	_ = notifier.Subscribe(event.TopicScanningCompleted, &p2p.Handler{})
	_ = notifier.Subscribe(event.TopicArtifactLabeled, &p2p.Handler{})

	// audit logs
	_ = notifier.Subscribe(event.TopicPushArtifact, &auditlog.Handler{})
	_ = notifier.Subscribe(event.TopicPullArtifact, &auditlog.Handler{})
	_ = notifier.Subscribe(event.TopicDeleteArtifact, &auditlog.Handler{})
	_ = notifier.Subscribe(event.TopicCreateProject, &auditlog.Handler{})
	_ = notifier.Subscribe(event.TopicDeleteProject, &auditlog.Handler{})
	_ = notifier.Subscribe(event.TopicDeleteRepository, &auditlog.Handler{})
	_ = notifier.Subscribe(event.TopicCreateTag, &auditlog.Handler{})
	_ = notifier.Subscribe(event.TopicDeleteTag, &auditlog.Handler{})

	// internal
	_ = notifier.Subscribe(event.TopicPullArtifact, &internal.Handler{})
	_ = notifier.Subscribe(event.TopicPushArtifact, &internal.Handler{})

	_ = task.RegisterTaskStatusChangePostFunc(job.Replication, func(ctx context.Context, taskID int64, status string) error {
		notification.AddEvent(ctx, &metadata.ReplicationMetaData{
			ReplicationTaskID: taskID,
			Status:            status,
		})
		return nil
	})
}
