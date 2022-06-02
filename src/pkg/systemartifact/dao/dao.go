package dao

import (
	"context"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/systemartifact/model"
)

const (
	sizeQuery = "select sum(size) as total_size from system_artifact"
)

// DAO defines an data access interface for manging the CRUD and read of system
// artifact tracking records
type DAO interface {

	// Create a system artifact tracking record.
	Create(ctx context.Context, systemArtifact *model.SystemArtifact) (int64, error)

	// Get a system artifact tracking record identified by vendor, repository and digest
	Get(ctx context.Context, vendor, repository, digest string) (*model.SystemArtifact, error)

	// Delete a system artifact tracking record identified by vendor, repository and digest
	Delete(ctx context.Context, vendor, repository, digest string) error

	// List all the system artifact records that match the criteria specified
	// within the query.
	List(ctx context.Context, query *q.Query) ([]*model.SystemArtifact, error)

	// Size returns the sum of all the system artifacts.
	Size(ctx context.Context) (int64, error)
}

// NewSystemArtifactDao returns an instance of the system artifact dao layer
func NewSystemArtifactDao() DAO {
	return &systemArtifactDAO{}
}

// The default implementation of the system artifact DAO.
type systemArtifactDAO struct{}

func (*systemArtifactDAO) Create(ctx context.Context, systemArtifact *model.SystemArtifact) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(systemArtifact)
	if err != nil {
		if e := orm.AsConflictError(err, "system artifact with repository name %s and digest %s already exists",
			systemArtifact.Repository, systemArtifact.Digest); e != nil {
			err = e
		}
		return int64(0), err
	}
	return id, nil
}

func (*systemArtifactDAO) Get(ctx context.Context, vendor, repository, digest string) (*model.SystemArtifact, error) {
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

func (*systemArtifactDAO) Delete(ctx context.Context, vendor, repository, digest string) error {
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

	return err
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

func (d *systemArtifactDAO) Size(ctx context.Context) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return int64(0), err
	}
	var totalSize int64
	if err := ormer.Raw(sizeQuery).QueryRow(&totalSize); err != nil {
		return int64(0), err
	}

	return totalSize, nil
}
