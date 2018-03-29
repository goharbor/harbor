package replication

import (
	"github.com/vmware/harbor/src/common/notifier"
	"github.com/vmware/harbor/src/replication/event/notification"
	"github.com/vmware/harbor/src/replication/event/topic"
)

//Task is the task for triggering one replication
type Task struct {
	PolicyID int64
}

//NewTask is constructor of creating ReplicationTask
func NewTask(policyID int64) *Task {
	return &Task{
		PolicyID: policyID,
	}
}

//Name returns the name of this task
func (t *Task) Name() string {
	return "replication"
}

//Run the actions here
func (t *Task) Run() error {
	return notifier.Publish(topic.StartReplicationTopic, notification.StartReplicationNotification{
		PolicyID: t.PolicyID,
	})
}
