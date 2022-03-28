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
	"context"
	"fmt"
	"time"

	"github.com/docker/distribution/manifest/schema2"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/blob/models"
)

// NewMysqlDao returns an instance of the mysql DAO
func NewMysqlDao() DAO {
	return &mysqlDao{}
}

type mysqlDao struct {
	*dao
}

func (d *mysqlDao) UpdateBlobStatus(ctx context.Context, blob *models.Blob) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return -1, err
	}

	var sql string
	if blob.Status == models.StatusNone {
		sql = "UPDATE `blob` SET version = version + 1, update_time = ?, status = ? where id = ? AND version >= ? AND status IN (%s)"
	} else {
		sql = "UPDATE `blob` SET version = version + 1, update_time = ?, status = ? where id = ? AND version = ? AND status IN (%s)"
	}

	var newVersion int64
	params := []interface{}{time.Now(), blob.Status, blob.ID, blob.Version}
	stats := models.StatusMap[blob.Status]
	for _, stat := range stats {
		params = append(params, stat)
	}

	if res, err := o.Raw(fmt.Sprintf(sql, orm.ParamPlaceholderForIn(len(models.StatusMap[blob.Status]))), params...).Exec(); err != nil {
		return -1, err
	} else if row, err := res.RowsAffected(); err == nil && row == 0 {
		log.Warningf("no blob is updated according to query condition, id: %d, status_in, %v", blob.ID, models.StatusMap[blob.Status])
		return 0, nil
	}

	selectVersionSQL := "SELECT version FROM `blob` WHERE id = ?"
	if err := o.Raw(selectVersionSQL, blob.ID).QueryRow(&newVersion); err != nil {
		return 0, nil
	}

	blob.Version = newVersion
	return 1, nil
}

func (d *mysqlDao) SumBlobsSizeByProject(ctx context.Context, projectID int64, excludeForeignLayer bool) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	params := []interface{}{projectID}
	sql := "SELECT SUM(size) FROM `blob` JOIN project_blob ON `blob`.id = project_blob.blob_id AND project_id = ?"
	if excludeForeignLayer {
		foreignLayerTypes := []interface{}{
			schema2.MediaTypeForeignLayer,
		}

		sql = fmt.Sprintf("%s AND content_type NOT IN (%s)", sql, orm.ParamPlaceholderForIn(len(foreignLayerTypes)))
		params = append(params, foreignLayerTypes...)
	}

	var totalSize int64
	if err := o.Raw(sql, params...).QueryRow(&totalSize); err != nil {
		return 0, err
	}

	return totalSize, nil
}

// SumBlobsSize returns sum size of all blobs skip foreign blobs when `excludeForeignLayer` is true
func (d *mysqlDao) SumBlobsSize(ctx context.Context, excludeForeignLayer bool) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	params := []interface{}{}

	sql := "SELECT SUM(size) FROM `blob`"
	if excludeForeignLayer {
		foreignLayerTypes := []interface{}{
			schema2.MediaTypeForeignLayer,
		}
		sql = fmt.Sprintf("%s Where content_type NOT IN (%s)", sql, orm.ParamPlaceholderForIn(len(foreignLayerTypes)))
		params = append(params, foreignLayerTypes...)
	}

	var totalSize int64
	if err := o.Raw(sql, params...).QueryRow(&totalSize); err != nil {
		return 0, err
	}

	return totalSize, nil
}

func (d *mysqlDao) ExistProjectBlob(ctx context.Context, projectID int64, blobDigest string) (bool, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return false, err
	}

	sql := "SELECT COUNT(*) FROM project_blob JOIN `blob` ON project_blob.blob_id = blob.id AND project_id = ? AND digest = ?"

	var count int64
	if err := o.Raw(sql, projectID, blobDigest).QueryRow(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (d *mysqlDao) GetBlobsNotRefedByProjectBlob(ctx context.Context, timeWindowHours int64) ([]*models.Blob, error) {
	var noneRefed []*models.Blob
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return noneRefed, err
	}

	sql := fmt.Sprintf("SELECT b.id, b.digest, b.content_type, b.status, b.version, b.size FROM `blob` AS b LEFT JOIN project_blob pb ON b.id = pb.blob_id WHERE pb.id IS NULL AND b.update_time <= date_sub(CURRENT_TIMESTAMP(6), interval %d hour);", timeWindowHours)
	_, err = ormer.Raw(sql).QueryRows(&noneRefed)
	if err != nil {
		return noneRefed, err
	}

	return noneRefed, nil
}

func (d *mysqlDao) GetBlobsByArtDigest(ctx context.Context, digest string) ([]*models.Blob, error) {
	var blobs []*models.Blob
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return blobs, err
	}

	sql := "SELECT b.id, b.digest, b.content_type, b.status, b.version, b.size FROM artifact_blob AS ab LEFT JOIN `blob` b ON ab.digest_blob = b.digest WHERE ab.digest_af = ?"
	_, err = ormer.Raw(sql, digest).QueryRows(&blobs)
	if err != nil {
		return blobs, err
	}

	return blobs, nil
}
