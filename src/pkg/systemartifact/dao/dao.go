package dao

import (
	"context"
	"github.com/goharbor/harbor/src/lib/orm"
)
import "github.com/goharbor/harbor/src/pkg/systemartifact/model"
import "github.com/goharbor/harbor/src/lib/q"

// DAO defines an data access interface for manging the CRUD and read of system
// artifact tracking records
type DAO interface {

	// Create a system artifact tracking record.
	Create(ctx context.Context, artifactRecord *model.SystemArtifact) (int64, error)

	// Get a system artifact tracking record identified by vendor, repository and digest
	Get(ctx context.Context, vendor string, repository string, digest string) (*model.SystemArtifact, error)

	// Delete a system artifact tracking record identified by vendor, repository and digest
	Delete(ctx context.Context, vendor string, repository string, digest string) error

	// List all the system artifact records that match the criteria specified
	// within the query.
	List(ctx context.Context, query *q.Query) ([]*model.SystemArtifact, error)
}

// NewSystemArtifactDao returns an instance of the system artifact dao layer
func NewSystemArtifactDao() DAO {
	return &systemArtifactDAO{}
}

// The default implementation of the system artifact DAO.
type systemArtifactDAO struct{}

func (*systemArtifactDAO) Create(ctx context.Context, artifactRecord *model.SystemArtifact) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(artifactRecord)
	if err != nil {
		if e := orm.AsConflictError(err, "system artifact with repository name %s and digest %s already exists",
			artifactRecord.Repository, artifactRecord.Digest); e != nil {
			err = e
		}
		return int64(0), err
	}
	return id, nil
}

func (*systemArtifactDAO) Get(ctx context.Context, vendor string, repository string, digest string) (*model.SystemArtifact, error) {
	ormer, err := orm.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	sa := model.SystemArtifact{Repository: repository, Digest: digest, Vendor: vendor}

	err = ormer.Read(&sa, "vendor", "repository", "digest")

	if err != nil {
		if e := orm.AsNotFoundError(err, "system artifact with repository name %s and digest %s not found",
			repository, digest); e != nil {
			err = e
		}
		return nil, err
	}

	return &sa, nil
}

func (*systemArtifactDAO) Delete(ctx context.Context, vendor string, repository string, digest string) error {
	ormer, err := orm.FromContext(ctx)

	if err != nil {
		return err
	}

	sa := model.SystemArtifact{
		Repository: repository,
		Digest:     digest,
		Vendor:     vendor,
	}

	_, err = ormer.Delete(&sa, "vendor", "repository", "digest")

	if err != nil {
		if e := orm.AsNotFoundError(err, "system artifact with repository name %s and digest %s not found",
			repository, digest); e != nil {
			err = e
		}
		return err
	}

	return nil
}

func (*systemArtifactDAO) List(ctx context.Context, query *q.Query) ([]*model.SystemArtifact, error) {
	qs, err := orm.QuerySetter(ctx, &model.SystemArtifact{}, query)

	if err != nil {
		return nil, err
	}
	var systemArtifactRecords []*model.SystemArtifact

	_, err = qs.All(&systemArtifactRecords)

	if err != nil {
		return nil, err
	}

	return systemArtifactRecords, nil
}
