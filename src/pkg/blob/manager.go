// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blob

import (
	"context"

	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/blob/dao"
	"github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/pkg/q"
)

// Blob alias `models.Blob` to make it natural to use the Manager
type Blob = models.Blob

// ListParams alias `models.ListParams` to make it natural to use the Manager
type ListParams = models.ListParams

var (
	// Mgr default blob manager
	Mgr = NewManager()
)

// Manager interface provide the management functions for blobs
type Manager interface {
	// AssociateWithArtifact associate blob with artifact
	AssociateWithArtifact(ctx context.Context, blobDigest, artifactDigest string) (int64, error)

	// AssociateWithProject associate blob with project
	AssociateWithProject(ctx context.Context, blobID, projectID int64) (int64, error)

	// Create create blob
	Create(ctx context.Context, digest string, contentType string, size int64) (int64, error)

	// CleanupAssociationsForArtifact remove all associations between blob and artifact by artifact digest
	CleanupAssociationsForArtifact(ctx context.Context, artifactDigest string) error

	// CleanupAssociationsForProject remove unneeded associations between blobs and project
	CleanupAssociationsForProject(ctx context.Context, projectID int64, blobs []*Blob) error

	// Get get blob by digest
	Get(ctx context.Context, digest string) (*Blob, error)

	// Update the blob
	Update(ctx context.Context, blob *Blob) error

	// List returns blobs by params
	List(ctx context.Context, params ListParams) ([]*Blob, error)

	// IsAssociatedWithArtifact returns true when blob associated with artifact
	IsAssociatedWithArtifact(ctx context.Context, blobDigest, artifactDigest string) (bool, error)

	// IsAssociatedWithProject returns true when blob associated with project
	IsAssociatedWithProject(ctx context.Context, digest string, projectID int64) (bool, error)
}

type manager struct {
	dao dao.DAO
}

func (m *manager) AssociateWithArtifact(ctx context.Context, blobDigest, artifactDigest string) (int64, error) {
	return m.dao.CreateArtifactAndBlob(ctx, artifactDigest, blobDigest)
}

func (m *manager) AssociateWithProject(ctx context.Context, blobID, projectID int64) (int64, error) {
	return m.dao.CreateProjectBlob(ctx, projectID, blobID)
}

func (m *manager) Create(ctx context.Context, digest string, contentType string, size int64) (int64, error) {
	return m.dao.CreateBlob(ctx, &Blob{Digest: digest, ContentType: contentType, Size: size})
}

func (m *manager) CleanupAssociationsForArtifact(ctx context.Context, artifactDigest string) error {
	return m.dao.DeleteArtifactAndBlobByArtifact(ctx, artifactDigest)
}

func (m *manager) CleanupAssociationsForProject(ctx context.Context, projectID int64, blobs []*Blob) error {
	if len(blobs) == 0 {
		return nil
	}

	shouldUnassociatedBlobs, err := m.dao.FindBlobsShouldUnassociatedWithProject(ctx, projectID, blobs)
	if err != nil {
		return err
	}

	var blobIDs []int64
	for _, blob := range shouldUnassociatedBlobs {
		blobIDs = append(blobIDs, blob.ID)
	}

	return m.dao.DeleteProjectBlob(ctx, projectID, blobIDs...)
}

func (m *manager) Get(ctx context.Context, digest string) (*Blob, error) {
	return m.dao.GetBlobByDigest(ctx, digest)
}

func (m *manager) Update(ctx context.Context, blob *Blob) error {
	return m.dao.UpdateBlob(ctx, blob)
}

func (m *manager) List(ctx context.Context, params ListParams) ([]*Blob, error) {
	kw := q.KeyWords{}

	if params.ArtifactDigest != "" {
		blobDigests, err := m.dao.GetAssociatedBlobDigestsForArtifact(ctx, params.ArtifactDigest)
		if err != nil {
			return nil, err
		}

		params.BlobDigests = append(params.BlobDigests, blobDigests...)
	}

	if len(params.BlobDigests) > 0 {
		kw["digest__in"] = params.BlobDigests
	}

	blobs, err := m.dao.ListBlobs(ctx, q.New(kw))
	if err != nil {
		return nil, err
	}

	var results []*Blob
	for _, blob := range blobs {
		results = append(results, blob)
	}

	return results, nil
}

func (m *manager) IsAssociatedWithArtifact(ctx context.Context, blobDigest, artifactDigest string) (bool, error) {
	md, err := m.dao.GetArtifactAndBlob(ctx, artifactDigest, blobDigest)
	if err != nil && !ierror.IsNotFoundErr(err) {
		return false, err
	}

	return md != nil, nil
}

func (m *manager) IsAssociatedWithProject(ctx context.Context, digest string, projectID int64) (bool, error) {
	return m.dao.ExistProjectBlob(ctx, projectID, digest)
}

// NewManager returns blob manager
func NewManager() Manager {
	return &manager{dao: dao.New()}
}
