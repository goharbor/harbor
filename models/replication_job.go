package models

import (
	"time"
)

const (
	//JobPending ...
	JobPending string = "pending"
	//JobRunning ...
	JobRunning string = "running"
	//JobError ...
	JobError string = "error"
	//JobStopped ...
	JobStopped string = "stopped"
	//JobFinished ...
	JobFinished string = "finished"
	//JobCanceled ...
	JobCanceled string = "canceled"
	//JobContinue is the status returned by statehandler to tell statemachine to move to next possible state based on trasition table.
	JobContinue string = "_continue"
	//RepOpTransfer represents the operation of a job to transfer repository to a remote registry/harbor instance.
	RepOpTransfer string = "transfer"
	//RepOpDelete represents the operation of a job to remove repository from a remote registry/harbor instance.
	RepOpDelete string = "delete"
	//UISecretCookie is the cookie name to contain the UI secret
	UISecretCookie string = "uisecret"
)

// RepPolicy is the model for a replication policy, which associate to a project and a target (destination)
type RepPolicy struct {
	ID        int64  `orm:"column(id)" json:"id"`
	ProjectID int64  `orm:"column(project_id)" json:"project_id"`
	TargetID  int64  `orm:"column(target_id)" json:"target_id"`
	Name      string `orm:"column(name)" json:"name"`
	//	Target       RepTarget `orm:"-" json:"target"`
	Enabled      int       `orm:"column(enabled)" json:"enabled"`
	Description  string    `orm:"column(description)" json:"description"`
	CronStr      string    `orm:"column(cron_str)" json:"cron_str"`
	StartTime    time.Time `orm:"column(start_time)" json:"start_time"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// RepJob is the model for a replication job, which is the execution unit on job service, currently it is used to transfer/remove
// a repository to/from a remote registry instance.
type RepJob struct {
	ID         int64    `orm:"column(id)" json:"id"`
	Status     string   `orm:"column(status)" json:"status"`
	Repository string   `orm:"column(repository)" json:"repository"`
	PolicyID   int64    `orm:"column(policy_id)" json:"policy_id"`
	Operation  string   `orm:"column(operation)" json:"operation"`
	Tags       string   `orm:"column(tags)" json:"-"`
	TagList    []string `orm:"-" json:"tags"`
	//	Policy       RepPolicy `orm:"-" json:"policy"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// RepTarget is the model for a replication targe, i.e. destination, which wraps the endpoint URL and username/password of a remote registry.
type RepTarget struct {
	ID           int64     `orm:"column(id)" json:"id"`
	URL          string    `orm:"column(url)" json:"url"`
	Name         string    `orm:"column(name)" json:"name"`
	Username     string    `orm:"column(username)" json:"username"`
	Password     string    `orm:"column(password)" json:"password"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

//TableName is required by by beego orm to map RepTarget to table replication_target
func (rt *RepTarget) TableName() string {
	return "replication_target"
}

//TableName is required by by beego orm to map RepJob to table replication_job
func (rj *RepJob) TableName() string {
	return "replication_job"
}

//TableName is required by by beego orm to map RepPolicy to table replication_policy
func (rp *RepPolicy) TableName() string {
	return "replication_policy"
}
