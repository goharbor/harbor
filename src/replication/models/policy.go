package models

import (
	"time"
)

//ReplicationPolicy defines the structure of a replication policy.
type ReplicationPolicy struct {
	ID                int64 //UUID of the policy
	Name              string
	Description       string
	Filters           []FilterItem
	ReplicateDeletion bool
	Trigger           *Trigger //The trigger of the replication
	ProjectIDs        []int64  //Projects attached to this policy
	TargetIDs         []int64
	CreationTime      time.Time
	UpdateTime        time.Time
}

//QueryParameter defines the parameters used to do query selection.
type QueryParameter struct {
	//Query by page, couple with pageSize
	Page int64

	//Size of each page, couple with page
	PageSize int64

	//Query by the name of trigger
	TriggerName string

	//Query by project ID
	ProjectID int64

	//Query by name
	Name string
}
