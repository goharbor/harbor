package dao

import (
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	liborm "github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
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
func ListHistoryRecords(query *q.Query) (int64, []*models.HistoryRecord, error) {
	qs, err := liborm.WithFilters(liborm.NewContext(nil, dao.GetOrmer()), &models.HistoryRecord{}, query)
	if err != nil {
		return 0, nil, err
	}

	total, err := qs.Count()
	if err != nil {
		return 0, nil, err
	}

	if query != nil {
		offset := (query.PageNumber - 1) * query.PageSize
		qs = qs.Offset(offset).Limit(query.PageSize)
	}

	var hrs []*models.HistoryRecord
	_, err = qs.All(&hrs)

	return total, hrs, err
}
