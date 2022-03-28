package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
)

// New returns an instance of the default DAO
func NewMysqlDao() DAO {
	return &mysqlDao{}
}

type mysqlDao struct {
	*dao
}

// Filter the results are: all of records in artifact_trash excludes the records in artifact with same repo and digest.
func (d *mysqlDao) Filter(ctx context.Context, cutOff time.Time) (arts []model.ArtifactTrash, err error) {
	var deletedAfs []model.ArtifactTrash
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return deletedAfs, err
	}

	sql := fmt.Sprintf(`SELECT aft.* FROM artifact_trash AS aft LEFT JOIN artifact af ON (aft.repository_name=af.repository_name AND aft.digest=af.digest) WHERE (af.digest IS NULL AND af.repository_name IS NULL) AND aft.creation_time <= FROM_UNIXTIME('%f')`, float64(cutOff.UnixNano())/float64((time.Second)))

	_, err = ormer.Raw(sql).QueryRows(&deletedAfs)
	if err != nil {
		return deletedAfs, err
	}

	return deletedAfs, nil
}

// Flush delete all of items beside the one in the time window.
func (d *mysqlDao) Flush(ctx context.Context, cutOff time.Time) (err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(`DELETE FROM artifact_trash where creation_time <= FROM_UNIXTIME('%f')`, float64(cutOff.UnixNano())/float64((time.Second)))
	if err != nil {
		return err
	}
	_, err = ormer.Raw(sql).Exec()
	if err != nil {
		return err
	}
	return nil
}
