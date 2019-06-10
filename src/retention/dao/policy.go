package dao

import (
	"errors"
	"fmt"

	"github.com/astaxie/beego/orm"
	commonDao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/retention/dao/models"
)

var deleteFilterQuery = fmt.Sprintf(`DELETE FROM %s WHERE policy = ?`, (&models.FilterMetadata{}).TableName())

// AddPolicy adds the specified policy to the database, returning the auto-generated ID.
//
// All associated filters are also persisted, within a single transaction.
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

// GetServerPolicy gets the server-wide retention policy from the database. Returns nil if no such policy exists.
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

// GetProjectPolicy gets the retention policy for the specified project ID. Returns nil if no such policy exists.
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

// GetRepoPolicy gets the retention policy for the specified repository ID. Returns nil if no such policy exists.
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

// UpdatePolicy updates the policy stored in the database. The policy must already be in the database.
//
// All associated filters are dropped and then re-added under a single transaction.
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

// DeletePolicy removes the specified policy from the database.
func DeletePolicy(id int64) error {
	_, err := commonDao.GetOrmer().Delete(&models.Policy{ID: id})
	return err
}
