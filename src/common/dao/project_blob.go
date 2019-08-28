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

package dao

import (
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common/models"
)

// AddBlobToProject ...
func AddBlobToProject(blobID, projectID int64) (int64, error) {
	pb := &models.ProjectBlob{
		BlobID:       blobID,
		ProjectID:    projectID,
		CreationTime: time.Now(),
	}

	_, id, err := GetOrmer().ReadOrCreate(pb, "blob_id", "project_id")
	return id, err
}

// AddBlobsToProject ...
func AddBlobsToProject(projectID int64, blobs ...*models.Blob) (int64, error) {
	if len(blobs) == 0 {
		return 0, nil
	}

	now := time.Now()

	var projectBlobs []*models.ProjectBlob
	for _, blob := range blobs {
		projectBlobs = append(projectBlobs, &models.ProjectBlob{
			BlobID:       blob.ID,
			ProjectID:    projectID,
			CreationTime: now,
		})
	}

	return GetOrmer().InsertMulti(len(projectBlobs), projectBlobs)
}

// RemoveBlobsFromProject ...
func RemoveBlobsFromProject(projectID int64, blobs ...*models.Blob) error {
	var blobIDs []interface{}
	for _, blob := range blobs {
		blobIDs = append(blobIDs, blob.ID)
	}

	if len(blobIDs) == 0 {
		return nil
	}

	sql := fmt.Sprintf(`DELETE FROM project_blob WHERE project_id = ? AND blob_id IN (%s)`, ParamPlaceholderForIn(len(blobIDs)))

	_, err := GetOrmer().Raw(sql, projectID, blobIDs).Exec()
	return err
}

// HasBlobInProject ...
func HasBlobInProject(projectID int64, digest string) (bool, error) {
	sql := `SELECT COUNT(*) FROM project_blob JOIN blob ON project_blob.blob_id = blob.id AND project_id = ? AND digest = ?`

	var count int64
	if err := GetOrmer().Raw(sql, projectID, digest).QueryRow(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetBlobsNotInProject returns blobs not in project
func GetBlobsNotInProject(projectID int64, blobDigests ...string) ([]*models.Blob, error) {
	if len(blobDigests) == 0 {
		return nil, nil
	}

	sql := fmt.Sprintf("SELECT * FROM blob WHERE id NOT IN (SELECT blob_id FROM project_blob WHERE project_id = ?) AND digest IN (%s)",
		ParamPlaceholderForIn(len(blobDigests)))

	params := []interface{}{projectID}
	for _, digest := range blobDigests {
		params = append(params, digest)
	}

	var blobs []*models.Blob
	if _, err := GetOrmer().Raw(sql, params...).QueryRows(&blobs); err != nil {
		return nil, err
	}

	return blobs, nil
}

// CountSizeOfProject ...
func CountSizeOfProject(pid int64) (int64, error) {
	var blobs []models.Blob

	sql := `
SELECT 
    DISTINCT bb.digest,
    bb.id,
    bb.content_type,
    bb.size,
    bb.creation_time
FROM artifact af
JOIN artifact_blob afnb
    ON af.digest = afnb.digest_af
JOIN BLOB bb
    ON afnb.digest_blob = bb.digest
WHERE af.project_id = ? 
`
	_, err := GetOrmer().Raw(sql, pid).QueryRows(&blobs)
	if err != nil {
		return 0, err
	}

	var size int64
	for _, blob := range blobs {
		size += blob.Size
	}

	return size, err
}

// RemoveUntaggedBlobs ...
func RemoveUntaggedBlobs(pid int64) error {
	var blobs []models.Blob
	sql := `
SELECT 
    DISTINCT bb.digest,
    bb.id,
    bb.content_type,
    bb.size,
    bb.creation_time
FROM artifact af
JOIN artifact_blob afnb
    ON af.digest = afnb.digest_af
JOIN BLOB bb
    ON afnb.digest_blob = bb.digest
WHERE af.project_id = ? 
`
	_, err := GetOrmer().Raw(sql, pid).QueryRows(&blobs)
	if len(blobs) == 0 {
		sql = fmt.Sprintf(`DELETE FROM project_blob WHERE project_id = ?`)
		_, err = GetOrmer().Raw(sql, pid).Exec()
		if err != nil {
			return err
		}
		return nil
	}

	var bbIDs []interface{}
	for _, bb := range blobs {
		bbIDs = append(bbIDs, bb.ID)
	}
	var projectBlobs []*models.ProjectBlob
	sql = fmt.Sprintf(`SELECT * FROM project_blob AS pb WHERE project_id = ? AND pb.blob_id NOT IN (%s)`, ParamPlaceholderForIn(len(bbIDs)))
	_, err = GetOrmer().Raw(sql, pid, bbIDs).QueryRows(&projectBlobs)
	if err != nil {
		return err
	}

	var pbIDs []interface{}
	for _, pb := range projectBlobs {
		pbIDs = append(pbIDs, pb.ID)
	}
	if len(pbIDs) == 0 {
		return nil
	}
	sql = fmt.Sprintf(`DELETE FROM project_blob WHERE id IN (%s)`, ParamPlaceholderForIn(len(pbIDs)))
	_, err = GetOrmer().Raw(sql, pbIDs).Exec()
	if err != nil {
		return err
	}

	return nil
}
