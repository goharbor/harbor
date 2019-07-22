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
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

// AddArtifactNBlob ...
func AddArtifactNBlob(afnb *models.ArtifactAndBlob) (int64, error) {
	now := time.Now()
	afnb.CreationTime = now
	id, err := GetOrmer().Insert(afnb)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return 0, ErrDupRows
		}
		return 0, err
	}
	return id, nil
}

// AddArtifactNBlobs ...
func AddArtifactNBlobs(afnbs []*models.ArtifactAndBlob) error {
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil {
		return err
	}

	var errInsertMultiple error
	total := len(afnbs)
	successNums, err := o.InsertMulti(total, afnbs)
	if err != nil {
		errInsertMultiple = err
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errInsertMultiple = errors.Wrap(errInsertMultiple, ErrDupRows.Error())
		}
		err := o.Rollback()
		if err != nil {
			log.Errorf("fail to rollback when to insert multiple artifact and blobs, %v", err)
			errInsertMultiple = errors.Wrap(errInsertMultiple, err.Error())
		}
		return errInsertMultiple
	}

	// part of them cannot be inserted successfully.
	if successNums != int64(total) {
		errInsertMultiple = errors.New("Not all of artifact and blobs are inserted successfully")
		err := o.Rollback()
		if err != nil {
			log.Errorf("fail to rollback when to insert multiple artifact and blobs, %v", err)
			errInsertMultiple = errors.Wrap(errInsertMultiple, err.Error())
		}
		return errInsertMultiple
	}

	err = o.Commit()
	if err != nil {
		log.Errorf("fail to commit when to insert multiple artifact and blobs, %v", err)
		return fmt.Errorf("fail to commit when to insert multiple artifact and blobs, %v", err)
	}

	return nil
}

// DeleteArtifactAndBlobByDigest ...
func DeleteArtifactAndBlobByDigest(digest string) error {
	_, err := GetOrmer().Raw(`delete from artifact_blob where digest_af = ? `, digest).Exec()
	if err != nil {
		return err
	}
	return nil
}

// CountSizeOfArtifact ...
func CountSizeOfArtifact(digest string) (int64, error) {
	var res []orm.Params
	num, err := GetOrmer().Raw(`SELECT sum(bb.size) FROM artifact_blob afnb LEFT JOIN blob bb ON afnb.digest_blob = bb.digest WHERE afnb.digest_af = ? `, digest).Values(&res)
	if err != nil {
		return -1, err
	}
	if num > 0 {
		size, err := strconv.ParseInt(res[0]["sum"].(string), 0, 64)
		if err != nil {
			return -1, err
		}
		return size, nil
	}
	return -1, err
}
