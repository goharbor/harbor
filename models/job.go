package models

import (
	"time"
)

type JobEntry struct {
	ID           int64                  `orm:"column(job_id)" json:"job_id"`
	Type         string                 `orm:"column(job_type)" json:"job_type"`
	OptionsStr   string                 `orm:"column(options)"`
	ParmsStr     string                 `orm:"column(parms)"`
	Status       string                 `orm:"column(status)" json:"status"`
	Options      map[string]interface{} `json:"options"`
	Parms        map[string]interface{} `json:"parms"`
	Enabled      int                    `orm:"column(enabled)" json:"enabled"`
	CronStr      string                 `orm:"column(cron_str)" json:"cron_str"`
	TriggeredBy  string                 `orm:"column(triggered_by)" json:"triggered_by"`
	CreationTime time.Time              `orm:"creation_time" json:"creation_time"`
	UpdateTime   time.Time              `orm:"update_time" json:"update_time"`
	Logs         []JobLog               `json:"logs"`
}

type JobLog struct {
	ID           int64     `orm:"column(log_id)" json:"log_id"`
	JobID        int64     `orm:"column(job_id)" json:"job_id"`
	Level        string    `orm:"column(level)" json:"level"`
	Message      string    `orm:"column(message)" json:"message"`
	CreationTime time.Time `orm:"creation_time" json:"creation_time"`
	UpdateTime   time.Time `orm:"update_time" json:"update_time"`
}
