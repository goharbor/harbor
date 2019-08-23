package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// AddBlob ...
func AddBlob(blob *models.Blob) (int64, error) {
	now := time.Now()
	blob.CreationTime = now
	id, err := GetOrmer().Insert(blob)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return 0, ErrDupRows
		}
		return 0, err
	}
	return id, nil
}

// GetOrCreateBlob returns blob by digest, create it if not exists
func GetOrCreateBlob(blob *models.Blob) (bool, *models.Blob, error) {
	blob.CreationTime = time.Now()

	created, id, err := GetOrmer().ReadOrCreate(blob, "digest")
	if err != nil {
		return false, nil, err
	}

	blob.ID = id

	return created, blob, nil
}

// GetBlob ...
func GetBlob(digest string) (*models.Blob, error) {
	o := GetOrmer()
	qs := o.QueryTable(&models.Blob{})
	qs = qs.Filter("Digest", digest)
	b := []*models.Blob{}
	_, err := qs.All(&b)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob for digest %s, error: %v", digest, err)
	}
	if len(b) == 0 {
		log.Infof("No blob found for digest %s, returning empty.", digest)
		return &models.Blob{}, nil
	} else if len(b) > 1 {
		log.Infof("Multiple blob found for digest %s", digest)
		return &models.Blob{}, fmt.Errorf("Multiple blob found for digest %s", digest)
	}
	return b[0], nil
}

// DeleteBlob ...
func DeleteBlob(digest string) error {
	o := GetOrmer()
	_, err := o.QueryTable("blob").Filter("digest", digest).Delete()
	return err
}

// GetBlobsByArtifact returns blobs of artifact
func GetBlobsByArtifact(artifactDigest string) ([]*models.Blob, error) {
	sql := `SELECT * FROM blob WHERE digest IN (SELECT digest_blob FROM artifact_blob WHERE digest_af = ?)`

	var blobs []*models.Blob
	if _, err := GetOrmer().Raw(sql, artifactDigest).QueryRows(&blobs); err != nil {
		return nil, err
	}

	return blobs, nil
}

// GetExclusiveBlobs returns layers of repository:tag which are not shared with other repositories in the project
func GetExclusiveBlobs(projectID int64, repository, digest string) ([]*models.Blob, error) {
	var exclusive []*models.Blob

	blobs, err := GetBlobsByArtifact(digest)
	if err != nil {
		return nil, err
	}
	if len(blobs) == 0 {
		return exclusive, nil
	}

	sql := fmt.Sprintf(`
SELECT
  DISTINCT b.digest_blob AS digest
FROM
  (
    SELECT
      digest
    FROM
      artifact
    WHERE
      (
        project_id = ?
        AND repo != ?
      )
      OR (
        project_id = ?
        AND digest != ?
      )
  ) AS a
  LEFT JOIN artifact_blob b ON a.digest = b.digest_af
  AND b.digest_blob IN (%s)`, ParamPlaceholderForIn(len(blobs)))

	params := []interface{}{projectID, repository, projectID, digest}
	for _, blob := range blobs {
		params = append(params, blob.Digest)
	}

	var rows []struct {
		Digest string
	}

	if _, err := GetOrmer().Raw(sql, params...).QueryRows(&rows); err != nil {
		return nil, err
	}

	shared := map[string]bool{}
	for _, row := range rows {
		shared[row.Digest] = true
	}

	for _, blob := range blobs {
		if !shared[blob.Digest] {
			exclusive = append(exclusive, blob)
		}
	}

	return exclusive, nil
}
