package task

import (
	"errors"
)

//ReplicationTask is the task for triggering one replication
type ReplicationTask struct{}

//NewReplicationTask is constructor of creating ReplicationTask
func NewReplicationTask() *ReplicationTask {
	return &ReplicationTask{}
}

//Name returns the name of this task
func (rt *ReplicationTask) Name() string {
	return "replication"
}

//Run the actions here
func (rt *ReplicationTask) Run() error {
	//Trigger the replication here
	return errors.New("Not implemented")
}
