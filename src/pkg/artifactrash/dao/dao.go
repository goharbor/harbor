package dao

import (
	"context"
	"time"

	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
)

// DAO is the data access object interface for artifact trash
type DAO interface {
	// Create the artifact trash
	Create(ctx context.Context, artifactrsh *model.ArtifactTrash) (id int64, err error)
	// Delete the artifact trash specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Filter lists the artifact that needs to be cleaned
	Filter(ctx context.Context) (arts []model.ArtifactTrash, err error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

// Create ...
func (d *dao) Create(ctx context.Context, artifactrsh *model.ArtifactTrash) (id int64, err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	artifactrsh.CreationTime = time.Now()
	id, err = ormer.Insert(artifactrsh)
	if err != nil {
		if e := orm.AsConflictError(err, "artifact trash %s already exists under the repository %s",
			artifactrsh.Digest, artifactrsh.RepositoryName); e != nil {
			err = e
		}
	}
	return id, err
}

// Delete ...
func (d *dao) Delete(ctx context.Context, id int64) (err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.ArtifactTrash{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return ierror.NotFoundError(nil).WithMessage("artifact trash %d not found", id)
	}
	return nil
}

// Filter ...
func (d *dao) Filter(ctx context.Context) (arts []model.ArtifactTrash, err error) {
	var deletedAfs []model.ArtifactTrash
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return deletedAfs, err
	}

	sql := `SELECT * FROM artifact_trash where artifact_trash.digest NOT IN (select digest from artifact)`

	if err := ormer.Raw(sql).QueryRow(&deletedAfs); err != nil {
		return deletedAfs, err
	}
	return deletedAfs, nil
}
