package notifier

import (
	"github.com/vmware/harbor/src/replication/event"
)

//Subscribe related topics
func init() {
	//Listen the related event topics
	Subscribe(event.StartReplicationTopic, &event.StartReplicationHandler{})
	Subscribe(event.ReplicationEventTopicOnPush, &event.OnPushHandler{})
	Subscribe(event.ReplicationEventTopicOnDeletion, &event.OnDeletionHandler{})
}
