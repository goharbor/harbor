package models

import (
	"time"
)

const (
	JobPending  string = "pending"
	JobRunning  string = "running"
	JobError    string = "error"
	JobStopped  string = "stopped"
	JobFinished string = "finished"
	JobCanceled string = "canceled"
	//  statemachine will move to next possible state based on trasition table
	JobContinue   string = "_continue"
	RepOpTransfer string = "transfer"
	RepOpDelete   string = "delete"
)

type RepPolicy struct {
	ID           int64     `orm:"column(id)" json:"id"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	TargetID     int64     `orm:"column(target_id)" json:"target_id"`
	Name         string    `orm:"column(name)" json:"name"`
	Target       RepTarget `orm:"-" json:"target"`
	Enabled      int       `orm:"column(enabled)" json:"enabled"`
	Description  string    `orm:"column(description)" json:"description"`
	CronStr      string    `orm:"column(cron_str)" json:"cron_str"`
	StartTime    time.Time `orm:"column(start_time)" json:"start_time"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

type RepJob struct {
	ID           int64     `orm:"column(id)" json:"id"`
	Status       string    `orm:"column(status)" json:"status"`
	Repository   string    `orm:"column(repository)" json:"repository"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	Operation    string    `orm:"column(operation)" json:"operation"`
	Policy       RepPolicy `orm:"-" json:"policy"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

type RepTarget struct {
	ID           int64     `orm:"column(id)" json:"id"`
	URL          string    `orm:"column(url)" json:"url"`
	Name         string    `orm:"column(name)" json:"name"`
	Username     string    `orm:"column(username)" json:"username"`
	Password     string    `orm:"column(password)" json:"password"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

func (rt *RepTarget) TableName() string {
	return "replication_target"
}

func (rj *RepJob) TableName() string {
	return "replication_job"
}

func (rp *RepPolicy) TableName() string {
	return "replication_policy"
}
