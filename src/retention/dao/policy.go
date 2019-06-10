package dao

import (
	"errors"
	"fmt"

	"github.com/astaxie/beego/orm"
	commonDao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/retention/dao/models"
)

var deleteFilterQuery = fmt.Sprintf(`DELETE FROM %s WHERE policy = ?`, (&models.FilterMetadata{}).TableName())

func AddPolicy(p *models.Policy) (id int64, err error) {
	o := commonDao.GetOrmer()
	if err = o.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			_ = o.Rollback()
		} else {
			err = o.Commit()
		}
	}()

	if id, err = o.Insert(p); err != nil {
		return
	}

	for _, f := range p.Filters {
		f.Policy = p

		if err = f.SyncJsonToORM(); err != nil {
			return
		}

		if _, err = o.Insert(f); err != nil {
			return
		}
	}

	return
}

func GetServerPolicy() (p *models.Policy, err error) {
	p = &models.Policy{}
	o := commonDao.GetOrmer()

	if err = o.
		QueryTable(p).
		Filter("project_id__isnull", true).
		Filter("repository_id__isnull", true).
		RelatedSel().
		One(p); err == orm.ErrNoRows {
		p = nil
		err = nil
	} else if err == nil {
		if _, err = o.LoadRelated(p, "Filters"); err != nil {
			return
		}

		for _, f := range p.Filters {
			if err = f.SyncORMToJson(); err != nil {
				return
			}
		}
	}

	return
}

func GetProjectPolicy(projectID int64) (p *models.Policy, err error) {
	p = &models.Policy{}
	o := commonDao.GetOrmer()

	if err = o.
		QueryTable(&models.Policy{}).
		Filter("project_id", projectID).
		Filter("repository_id__isnull", true).
		RelatedSel().
		One(p); err == orm.ErrNoRows {
		p = nil
		err = nil
	} else if err == nil {
		if _, err = o.LoadRelated(p, "Filters"); err != nil {
			return
		}

		for _, f := range p.Filters {
			if err = f.SyncORMToJson(); err != nil {
				return
			}
		}
	}

	return
}

func GetRepoPolicy(projectID, repoID int64) (p *models.Policy, err error) {
	p = &models.Policy{}
	o := commonDao.GetOrmer()

	if err = o.
		QueryTable(p).
		Filter("project_id", projectID).
		Filter("repository_id", repoID).
		RelatedSel().
		One(p); err == orm.ErrNoRows {
		p = nil
		err = nil
	} else if err == nil {
		if _, err = o.LoadRelated(p, "Filters"); err != nil {
			return
		}

		for _, f := range p.Filters {
			if err = f.SyncORMToJson(); err != nil {
				return
			}
		}
	}

	return
}

func UpdatePolicy(p *models.Policy, props ...string) (err error) {
	o := commonDao.GetOrmer()
	if err = o.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			_ = o.Rollback()
		} else {
			err = o.Commit()
		}
	}()

	if p == nil {
		err = errors.New("policy: cannot be nil")
		return
	}

	_, err = o.Update(p, props...)
	if err != nil {
		return
	}

	// Easiest way to update filter metadata is to just drop and re-add all of them
	if _, err = o.Raw(deleteFilterQuery, p.ID).Exec(); err != nil {
		return
	}

	for _, f := range p.Filters {
		f.Policy = p

		if err = f.SyncJsonToORM(); err != nil {
			return
		}

		if _, err = o.Insert(f); err != nil {
			return
		}
	}

	return
}

func DeletePolicy(id int64) error {
	_, err := commonDao.GetOrmer().Delete(&models.Policy{ID: id})
	return err
}
