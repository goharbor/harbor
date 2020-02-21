package dao

import (
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
)

// ListHistoryQuery defines the query params of the history record.
type ListHistoryQuery struct {
	Page     uint
	PageSize uint
	Keyword  string
}

// AddHistoryRecord adds one distribution history record.
func AddHistoryRecord(history *models.HistoryRecord) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(history)
}

// GetHistoryRecord gets distribution history record by id.
func GetHistoryRecord(id int64) (*models.HistoryRecord, error) {
	o := dao.GetOrmer()
	dhr := models.HistoryRecord{ID: id}
	err := o.Read(&dhr)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &dhr, err
}

// GetHistoryRecordByTaskID gets distribution history record by TaskID.
func GetHistoryRecordByTaskID(taskID string) (*models.HistoryRecord, error) {
	o := dao.GetOrmer()
	dhr := models.HistoryRecord{TaskID: taskID}
	err := o.Read(&dhr, "TaskID")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &dhr, err
}

// UpdateHistoryRecord updates distribution history record.
func UpdateHistoryRecord(history *models.HistoryRecord, props ...string) error {
	o := dao.GetOrmer()
	_, err := o.Update(history, props...)
	return err
}

// DeleteHistoryRecord deletes one distribution history record by id.
func DeleteHistoryRecord(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.HistoryRecord{ID: id})
	return err
}

// ListHistoryRecords lists history records by query parmas.
func ListHistoryRecords(query *ListHistoryQuery) ([]*models.HistoryRecord, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(&models.HistoryRecord{})

	if query != nil {
		offset := (query.Page - 1) * query.PageSize
		qs = qs.Offset(offset).Limit(query.PageSize)
		// keyword match
		if len(query.Keyword) > 0 {
			qs = qs.Filter("image__contains", query.Keyword)
		}
	}

	var hrs []*models.HistoryRecord
	_, err := qs.All(&hrs)
	return hrs, err
}
