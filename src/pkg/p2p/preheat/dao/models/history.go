package models

// HistoryRecord defines distribution history record.
type HistoryRecord struct {
	ID         int64  `orm:"pk;auto;column(id)" json:"id"`
	TaskID     string `orm:"column(task_id)" json:"task_id"`
	Image      string `orm:"column(image)" json:"image"`
	StartTime  string `orm:"column(start_time)" json:"start_time"`
	FinishTime string `orm:"column(finish_time)" json:"finish_time"`
	Status     string `orm:"column(status)" json:"status"`
	Provider   string `orm:"column(provider)" json:"provider"`
	Instance   int64  `orm:"column(instance)" json:"instance"`
}

// TableName set table name for ORM.
func (hr *HistoryRecord) TableName() string {
	return "p2p_preheat_history"
}
