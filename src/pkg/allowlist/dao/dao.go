package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	beegoorm "github.com/beego/beego/orm"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/allowlist/models"
)

// DAO is the data access object interface for CVE allowlist
type DAO interface {
	// Set creates or updates the CVE allowlist to DB based on the project ID in the input parm, if the project does not
	// have a CVE allowlist, an empty allowlist will be created.  The project ID should be 0 for system level CVE allowlist
	Set(ctx context.Context, l models.CVEAllowlist) (int64, error)
	// QueryByProjectID returns the CVE allowlist of the project based on the project ID in parameter.  The project ID should be 0
	// for system level CVE allowlist
	QueryByProjectID(ctx context.Context, pid int64) (*models.CVEAllowlist, error)
}

// New ...
func New() DAO {
	return &dao{}
}

func init() {
	beegoorm.RegisterModel(new(models.CVEAllowlist))
}

type dao struct{}

func (d *dao) Set(ctx context.Context, l models.CVEAllowlist) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	l.CreationTime = now
	l.UpdateTime = now
	itemsBytes, _ := json.Marshal(l.Items)
	l.ItemsText = string(itemsBytes)
	return ormer.InsertOrUpdate(&l, "project_id")
}

func (d *dao) QueryByProjectID(ctx context.Context, pid int64) (*models.CVEAllowlist, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	qs := ormer.QueryTable(&models.CVEAllowlist{})
	qs = qs.Filter("ProjectID", pid)
	var r []models.CVEAllowlist
	_, err = qs.All(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to get CVE allowlist for project %d, error: %v", pid, err)
	}
	if len(r) == 0 {
		return nil, nil
	} else if len(r) > 1 {
		log.Infof("Multiple CVE allowlists found for project %d, length: %d, returning first element.", pid, len(r))
	}
	items := []models.CVEAllowlistItem{}
	err = json.Unmarshal([]byte(r[0].ItemsText), &items)
	if err != nil {
		log.Errorf("Failed to decode item list, err: %v, text: %s", err, r[0].ItemsText)
		return nil, err
	}
	r[0].Items = items
	return &r[0], nil
}
