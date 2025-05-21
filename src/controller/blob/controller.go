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
	"fmt"
	"time"

	"github.com/docker/distribution"
	"github.com/go-redis/redis/v8"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	libredis "github.com/goharbor/harbor/src/lib/redis"
	"github.com/goharbor/harbor/src/pkg/blob"
	blob_models "github.com/goharbor/harbor/src/pkg/blob/models"
)

var (
	// Ctl is a global blob controller instance
	Ctl = NewController()
)

// Controller defines the operations related with blobs
type Controller interface {
	// AssociateWithArtifact associate blobs with manifest.
	AssociateWithArtifact(ctx context.Context, blobDigests []string, artifactDigest string) error

	// AssociateWithProjectByID associate blob with project by blob id
	AssociateWithProjectByID(ctx context.Context, blobID int64, projectID int64) error

	// AssociateWithProjectByDigest associate blob with project by blob digest
	AssociateWithProjectByDigest(ctx context.Context, blobDigest string, projectID int64) error

	// CalculateTotalSizeByProject returns the sum of the blob size for the project
	CalculateTotalSizeByProject(ctx context.Context, projectID int64, excludeForeign bool) (int64, error)

	// CalculateTotalSize returns the sum of all the blobs size
	CalculateTotalSize(ctx context.Context, excludeForeign bool) (int64, error)

	// Ensure create blob when it not exist.
	Ensure(ctx context.Context, digest string, contentType string, size int64) (int64, error)

	// Exist check blob exist by digest,
	// it check the blob associated with the artifact when `IsAssociatedWithArtifact` option provided,
	// and also check the blob associated with the project when `IsAssociatedWithProject` option provied.
	Exist(ctx context.Context, digest string, options ...Option) (bool, error)

	// FindMissingAssociationsForProjectByArtifact returns blobs which are associated with artifact but not associated with project
	FindMissingAssociationsForProject(ctx context.Context, projectID int64, blobs []*blob.Blob) ([]*blob.Blob, error)

	// Get get the blob by digest,
	// it check the blob associated with the artifact when `IsAssociatedWithArtifact` option provided,
	// and also check the blob associated with the project when `IsAssociatedWithProject` option provied.
	Get(ctx context.Context, digest string, options ...Option) (*blob.Blob, error)

	// List list blobs
	List(ctx context.Context, query *q.Query) ([]*blob.Blob, error)

	// Sync create blobs from `References` when they are not exist
	// and update the blob content type when they are exist,
	Sync(ctx context.Context, references []distribution.Descriptor) error

	// SetAcceptedBlobSize update the accepted size of stream upload blob.
	SetAcceptedBlobSize(ctx context.Context, sessionID string, size int64) error

	// GetAcceptedBlobSize returns the accepted size of stream upload blob.
	GetAcceptedBlobSize(ctx context.Context, sessionID string) (int64, error)

	// Touch updates the blob status to StatusNone and increase version every time.
	Touch(ctx context.Context, blob *blob.Blob) error

	// Fail updates the blob status to StatusDeleteFailed and increase version every time.
	Fail(ctx context.Context, blob *blob.Blob) error

	// Update updates the blob, it cannot handle blob status transitions.
	Update(ctx context.Context, blob *blob.Blob) error

	// Delete deletes the blob by its id
	Delete(ctx context.Context, id int64) error
}

// NewController creates an instance of the default repository controller
func NewController() Controller {
	return &controller{
		blobMgr:            blob.Mgr,
		blobSizeExpiration: time.Hour * 24, // keep the size of blob in redis with 24 hours
	}
}

type controller struct {
	blobMgr            blob.Manager
	blobSizeExpiration time.Duration
}

func (c *controller) AssociateWithArtifact(ctx context.Context, blobDigests []string, artifactDigest string) error {
	exist, err := c.Exist(ctx, artifactDigest, IsAssociatedWithArtifact(artifactDigest))
	if err != nil {
		return err
	}

	if exist {
		log.G(ctx).Infof("artifact digest %s already exist, skip to associate blobs with the artifact", artifactDigest)
		return nil
	}

	for _, blobDigest := range append(blobDigests, artifactDigest) {
		_, err := c.blobMgr.AssociateWithArtifact(ctx, blobDigest, artifactDigest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *controller) AssociateWithProjectByID(ctx context.Context, blobID int64, projectID int64) error {
	_, err := c.blobMgr.AssociateWithProject(ctx, blobID, projectID)
	return err
}

func (c *controller) AssociateWithProjectByDigest(ctx context.Context, blobDigest string, projectID int64) error {
	blob, err := c.blobMgr.Get(ctx, blobDigest)
	if err != nil {
		return err
	}

	_, err = c.blobMgr.AssociateWithProject(ctx, blob.ID, projectID)
	return err
}

func (c *controller) CalculateTotalSizeByProject(ctx context.Context, projectID int64, excludeForeign bool) (int64, error) {
	return c.blobMgr.CalculateTotalSizeByProject(ctx, projectID, excludeForeign)
}

func (c *controller) CalculateTotalSize(ctx context.Context, excludeForeign bool) (int64, error) {
	return c.blobMgr.CalculateTotalSize(ctx, excludeForeign)
}

func (c *controller) Ensure(ctx context.Context, digest string, contentType string, size int64) (blobID int64, err error) {
	blob, err := c.blobMgr.Get(ctx, digest)
	if err == nil {
		return blob.ID, nil
	}

	if !errors.IsNotFoundErr(err) {
		return 0, err
	}

	return c.blobMgr.Create(ctx, digest, contentType, size)
}

func (c *controller) Exist(ctx context.Context, digest string, options ...Option) (bool, error) {
	if digest == "" {
		return false, errors.BadRequestError(nil).WithMessage("exist blob require digest")
	}

	_, err := c.Get(ctx, digest, options...)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c *controller) FindMissingAssociationsForProject(ctx context.Context, projectID int64, blobs []*blob.Blob) ([]*blob.Blob, error) {
	if len(blobs) == 0 {
		return nil, nil
	}

	var ol q.OrList
	for _, blob := range blobs {
		ol.Values = append(ol.Values, blob.Digest)
	}

	associatedBlobs, err := c.blobMgr.List(ctx, q.New(q.KeyWords{"digest": &ol, "projectID": projectID}))
	if err != nil {
		return nil, err
	}

	if len(associatedBlobs) == 0 {
		return blobs, nil
	} else if len(associatedBlobs) == len(blobs) {
		return nil, nil
	}

	associated := make(map[string]bool, len(associatedBlobs))
	for _, blob := range associatedBlobs {
		associated[blob.Digest] = true
	}

	var results []*blob.Blob
	for _, blob := range blobs {
		if !associated[blob.Digest] {
			results = append(results, blob)
		}
	}

	return results, nil
}

func (c *controller) Get(ctx context.Context, digest string, options ...Option) (*blob.Blob, error) {
	if digest == "" {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("require digest")
	}

	opts := newOptions(options...)

	keywords := make(map[string]any)
	if digest != "" {
		ol := q.OrList{
			Values: []any{
				digest,
			},
		}
		keywords["digest"] = &ol
	}
	if opts.ProjectID != 0 {
		keywords["projectID"] = opts.ProjectID
	}
	if opts.ArtifactDigest != "" {
		keywords["artifactDigest"] = opts.ArtifactDigest
	}
	query := &q.Query{
		Keywords: keywords,
	}

	blobs, err := c.blobMgr.List(ctx, query)
	if err != nil {
		return nil, err
	} else if len(blobs) == 0 {
		return nil, errors.NotFoundError(nil).WithMessagef("blob %s not found", digest)
	}

	return blobs[0], nil
}

func (c *controller) List(ctx context.Context, query *q.Query) ([]*blob.Blob, error) {
	return c.blobMgr.List(ctx, query)
}

func (c *controller) Sync(ctx context.Context, references []distribution.Descriptor) error {
	if len(references) == 0 {
		return nil
	}

	var ol q.OrList
	for _, reference := range references {
		ol.Values = append(ol.Values, reference.Digest.String())
	}

	blobs, err := c.blobMgr.List(ctx, q.New(q.KeyWords{"digest": &ol}))
	if err != nil {
		return err
	}

	mp := make(map[string]*blob.Blob, len(blobs))
	for _, blob := range blobs {
		mp[blob.Digest] = blob
	}

	var missing, updating []*blob.Blob
	for _, reference := range references {
		if exist, found := mp[reference.Digest.String()]; found {
			if exist.ContentType != reference.MediaType {
				exist.ContentType = reference.MediaType
				updating = append(updating, exist)
			}
		} else {
			missing = append(missing, &blob.Blob{
				Digest:      reference.Digest.String(),
				ContentType: reference.MediaType,
				Size:        reference.Size,
			})
		}
	}

	if len(updating) > 0 {
		err := orm.WithTransaction(func(ctx context.Context) error {
			for _, blob := range updating {
				if err := c.Update(ctx, blob); err != nil {
					log.G(ctx).Warningf("Failed to update blob %s, error: %v", blob.Digest, err)
					return err
				}
			}

			return nil
		})(orm.SetTransactionOpNameToContext(ctx, "tx-sync-blob"))
		if err != nil {
			return err
		}
	}

	if len(missing) > 0 {
		for _, blob := range missing {
			if _, err := c.blobMgr.Create(ctx, blob.Digest, blob.ContentType, blob.Size); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *controller) SetAcceptedBlobSize(ctx context.Context, sessionID string, size int64) error {
	key := blobSizeKey(sessionID)
	rc, err := libredis.GetRegistryClient()
	if err != nil {
		return err
	}

	err = rc.Set(ctx, key, size, c.blobSizeExpiration).Err()
	if err != nil {
		log.Errorf("failed to set accepted blob size for session %s in redis, error: %v", sessionID, err)
		return err
	}

	return nil
}

func (c *controller) GetAcceptedBlobSize(ctx context.Context, sessionID string) (int64, error) {
	key := blobSizeKey(sessionID)
	rc, err := libredis.GetRegistryClient()
	if err != nil {
		return 0, err
	}

	size, err := rc.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}

		return 0, err
	}

	return size, nil
}

func (c *controller) Touch(ctx context.Context, blob *blob.Blob) error {
	blob.Status = blob_models.StatusNone
	count, err := c.blobMgr.UpdateBlobStatus(ctx, blob)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New(nil).WithMessagef("no blob item is updated to StatusNone, id:%d, digest:%s", blob.ID, blob.Digest).WithCode(errors.NotFoundCode)
	}
	return nil
}

func (c *controller) Fail(ctx context.Context, blob *blob.Blob) error {
	blob.Status = blob_models.StatusDeleteFailed
	count, err := c.blobMgr.UpdateBlobStatus(ctx, blob)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New(nil).WithMessagef("no blob item is updated to StatusDeleteFailed, id:%d, digest:%s", blob.ID, blob.Digest).WithCode(errors.NotFoundCode)
	}
	return nil
}

func (c *controller) Update(ctx context.Context, blob *blob.Blob) error {
	return c.blobMgr.Update(ctx, blob)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	return c.blobMgr.Delete(ctx, id)
}

func blobSizeKey(sessionID string) string {
	return fmt.Sprintf("upload:%s:size", sessionID)
}
