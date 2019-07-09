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
	"github.com/goharbor/harbor/src/common/models"
	"strings"
	"time"
)

// AddArtifact ...
func AddArtifact(af *models.Artifact) (int64, error) {
	now := time.Now()
	af.CreationTime = now
	id, err := GetOrmer().Insert(af)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return 0, ErrDupRows
		}
		return 0, err
	}
	return id, nil
}

// DeleteArtifact ...
func DeleteArtifact(id int64) error {
	_, err := GetOrmer().QueryTable(&models.Artifact{}).Filter("ID", id).Delete()
	return err
}

// DeleteArtifactByDigest ...
func DeleteArtifactByDigest(digest string) error {
	_, err := GetOrmer().Raw(`delete from artifact where digest = ? `, digest).Exec()
	if err != nil {
		return err
	}
	return nil
}

// DeleteByTag ...
func DeleteByTag(projectID int, repo, tag string) error {
	_, err := GetOrmer().Raw(`delete from artifact where project_id = ? and repo = ? and tag = ? `,
		projectID, repo, tag).Exec()
	if err != nil {
		return err
	}
	return nil
}
