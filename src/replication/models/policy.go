package models

//ReplicationPolicy defines the structure of a replication policy.
type ReplicationPolicy struct {
	//UUID of the policy
	ID int

	//Projects attached to this policy
	RelevantProjects []int

	//The trigger of the replication
	Trigger Trigger
}

//QueryParameter defines the parameters used to do query selection.
type QueryParameter struct {
	//Query by page, couple with pageSize
	Page int

	//Size of each page, couple with page
	PageSize int

	//Query by the name of trigger
	TriggerName string

	//Query by project ID
	ProjectID int
}
