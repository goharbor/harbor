package systemartifact

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/pkg/systemartifact/dao"
	"github.com/goharbor/harbor/src/pkg/systemartifact/model"
	"io"
	"sync"
)

var (
	Mgr       = NewManager()
	keyFormat = "%s:%s"
)

const repositoryFormat = "sys_harbor/%s/%s"

// Manager provides a low-level interface for harbor services
// to create registry artifacts containing arbitrary data but which
// are not standard OCI artifacts.
// By using this framework, harbor components can create artifacts for
// cross component data sharing. The framework abstracts out the book-keeping
// logic involved in managing and tracking system artifacts.
// The Manager ultimately relies on the harbor registry client to perform
// the BLOB related operations into the registry.
type Manager interface {

	// Create a system artifact described by artifact record.
	// The reader would be used to read from the underlying data artifact.
	// Returns a system artifact tracking record id or any errors encountered in the data artifact upload process.
	// Invoking this API would result in a repository being created with the specified name and digest within the registry.
	Create(ctx context.Context, artifactRecord *model.SystemArtifact, reader io.Reader) (int64, error)

	// Read a system artifact described by repository name and digest.
	// The reader is responsible for closing the IO stream after the read completes.
	Read(ctx context.Context, vendor string, repository string, digest string) (io.ReadCloser, error)

	// Delete deletes a system artifact identified by a repository name and digest.
	// Also deletes the tracking record from the underlying table.
	Delete(ctx context.Context, vendor string, repository string, digest string) error

	// Exists checks for the existence of a system artifact identified by repository and digest.
	// A system artifact is considered as in existence if both the following conditions are true:
	// 1. There is a system artifact tracking record within the Harbor DB
	// 2. There is a BLOB corresponding to the repository name and digest obtained from system artifact record.
	Exists(ctx context.Context, vendor string, repository string, digest string) (bool, error)

	// GetStorageSize returns the total disk space used by the system artifacts stored in the registry.
	GetStorageSize(ctx context.Context) (int64, error)

	// RegisterCleanupCriteria a clean-up criteria for a specific vendor and artifact type combination.
	RegisterCleanupCriteria(vendor string, artifactType string, criteria Selector)

	// GetCleanupCriteria returns a clean-up criteria for a specific vendor and artifact type combination.
	// if no clean-up criteria is found then the default clean-up criteria is returned
	GetCleanupCriteria(vendor string, artifactType string) Selector

	// Cleanup cleans up the system artifacts (tracking records as well as blobs) based on the
	// artifact records selected by the Selector registered for each vendor type.
	// Returns the total number of records deleted, the reclaimed size and any error (if encountered)
	Cleanup(ctx context.Context) (int64, int64, error)
}

type systemArtifactManager struct {
	regCli                  registry.Client
	dao                     dao.DAO
	defaultCleanupCriterion Selector
	cleanupCriteria         map[string]Selector
	lock                    sync.Mutex
}

func NewManager() Manager {
	sysArtifactMgr := &systemArtifactManager{
		regCli:                  registry.Cli,
		dao:                     dao.NewSystemArtifactDao(),
		defaultCleanupCriterion: DefaultSelector,
		cleanupCriteria:         make(map[string]Selector),
	}
	return sysArtifactMgr
}

func (mgr *systemArtifactManager) Create(ctx context.Context, artifactRecord *model.SystemArtifact, reader io.Reader) (int64, error) {

	var artifactId int64

	// the entire create operation is executed within a transaction to ensure that any failures
	// during the blob creation or tracking record creation result in a rollback of the transaction
	createError := orm.WithTransaction(func(ctx context.Context) error {
		id, err := mgr.dao.Create(ctx, artifactRecord)
		if err != nil {
			log.Errorf("Error creating system artifact record for %s/%s/%s: %v", artifactRecord.Vendor, artifactRecord.Repository, artifactRecord.Digest, err)
			return err
		}
		repoName := mgr.getRepositoryName(artifactRecord.Vendor, artifactRecord.Repository)
		err = mgr.regCli.PushBlob(repoName, artifactRecord.Digest, artifactRecord.Size, reader)
		if err != nil {
			return err
		}
		artifactId = id
		return nil
	})(ctx)

	return artifactId, createError
}

func (mgr *systemArtifactManager) Read(ctx context.Context, vendor string, repository string, digest string) (io.ReadCloser, error) {
	sa, err := mgr.dao.Get(ctx, vendor, repository, digest)
	if err != nil {
		return nil, err
	}
	repoName := mgr.getRepositoryName(vendor, repository)
	_, readCloser, err := mgr.regCli.PullBlob(repoName, sa.Digest)
	if err != nil {
		return nil, err
	}
	return readCloser, nil
}

func (mgr *systemArtifactManager) Delete(ctx context.Context, vendor string, repository string, digest string) error {

	repoName := mgr.getRepositoryName(vendor, repository)
	if err := mgr.regCli.DeleteBlob(repoName, digest); err != nil {
		log.Errorf("Error deleting system artifact BLOB : %s. Error: %v", repoName, err)
		return err
	}

	return mgr.dao.Delete(ctx, vendor, repository, digest)
}

func (mgr *systemArtifactManager) Exists(ctx context.Context, vendor string, repository string, digest string) (bool, error) {
	_, err := mgr.dao.Get(ctx, vendor, repository, digest)
	if err != nil {
		return false, err
	}

	repoName := mgr.getRepositoryName(vendor, repository)
	exist, err := mgr.regCli.BlobExist(repoName, digest)

	if err != nil {
		return false, err
	}

	return exist, nil
}

func (mgr *systemArtifactManager) GetStorageSize(ctx context.Context) (int64, error) {
	return mgr.dao.Size(ctx)
}

func (mgr *systemArtifactManager) RegisterCleanupCriteria(vendor string, artifactType string, criteria Selector) {
	key := fmt.Sprintf(keyFormat, vendor, artifactType)
	defer mgr.lock.Unlock()
	mgr.lock.Lock()
	mgr.cleanupCriteria[key] = criteria
}

func (mgr *systemArtifactManager) GetCleanupCriteria(vendor string, artifactType string) Selector {
	key := fmt.Sprintf(keyFormat, vendor, artifactType)
	defer mgr.lock.Unlock()
	mgr.lock.Lock()
	if criteria, ok := mgr.cleanupCriteria[key]; ok {
		return criteria
	}
	return DefaultSelector
}

func (mgr *systemArtifactManager) Cleanup(ctx context.Context) (int64, int64, error) {
	logger.Info("Starting system artifact cleanup")
	// clean up artifact records having customized cleanup criteria first
	totalReclaimedSize := int64(0)
	totalRecordsDeleted := int64(0)

	// get a copy of the registered cleanup criteria and
	// iterate through this copy to invoke the cleanup
	registeredCriteria := make(map[string]Selector, 0)
	mgr.lock.Lock()
	for key, val := range mgr.cleanupCriteria {
		registeredCriteria[key] = val
	}
	mgr.lock.Unlock()

	for key, val := range registeredCriteria {
		logger.Infof("Executing cleanup for 'vendor:artifactType' : %s", key)
		deleted, size, err := mgr.cleanup(ctx, val)
		totalRecordsDeleted += deleted
		totalReclaimedSize += size

		if err != nil {
			// one vendor error should not impact the clean-up of other vendor types. Hence the cleanup logic would continue
			// after logging the error
			logger.Errorf("Error when cleaning up system artifacts for 'vendor:artifactType':%s, %v", key, err)
		}

	}

	logger.Info("Executing cleanup for default cleanup criteria")
	// clean up artifact records using the default criteria
	deleted, size, err := mgr.cleanup(ctx, mgr.defaultCleanupCriterion)
	if err != nil {
		// one vendor error should not impact the clean-up of other vendor types. Hence the cleanup logic would continue
		// after logging the error
		logger.Errorf("Error when cleaning up system artifacts for 'vendor:artifactType':%s, %v", "DefaultCriteria", err)
	}
	totalRecordsDeleted += deleted
	totalReclaimedSize += size

	return totalRecordsDeleted, totalReclaimedSize, nil
}

func (mgr *systemArtifactManager) cleanup(ctx context.Context, criteria Selector) (int64, int64, error) {
	// clean up artifact records having customized cleanup criteria first
	totalReclaimedSize := int64(0)
	totalRecordsDeleted := int64(0)

	isDefaultSelector := criteria == mgr.defaultCleanupCriterion

	records, err := criteria.List(ctx)

	if err != nil {

		return totalRecordsDeleted, totalReclaimedSize, err
	}

	for _, record := range records {
		// skip vendor artifact types with custom clean-up criteria registered
		if isDefaultSelector && mgr.isSelectorRegistered(record.Vendor, record.Type) {
			continue
		}
		err = mgr.Delete(ctx, record.Vendor, record.Repository, record.Digest)
		if err != nil {
			logger.Errorf("Error cleaning up artifact record for vendor: %s, repository: %s, digest: %s", record.Vendor, record.Repository, record.Digest)
			return totalRecordsDeleted, totalReclaimedSize, err
		}
		totalReclaimedSize += record.Size
		totalRecordsDeleted += 1
	}
	return totalRecordsDeleted, totalReclaimedSize, nil
}

func (mgr *systemArtifactManager) getRepositoryName(vendor string, repository string) string {
	return fmt.Sprintf(repositoryFormat, vendor, repository)
}

func (mgr *systemArtifactManager) isSelectorRegistered(vendor, artifactType string) bool {
	key := fmt.Sprintf(keyFormat, vendor, artifactType)
	_, ok := mgr.cleanupCriteria[key]
	return ok
}
