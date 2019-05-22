// Copyright 2019 Project Harbor Authors
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

package filter

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"

	"github.com/goharbor/harbor/src/common/retention"
)

const (
	FilterTypeDeleteOlderThan = "retention:filter:delete_older_than"
)

var daysAgo = -24 * time.Hour

type deleteOlderThan struct {
	n int
}

func NewDeleteOlderThan(metadata map[string]interface{}) (*deleteOlderThan, error) {
	if raw, ok := metadata[MetaDataKeyN]; ok {
		if n, ok := raw.(int); ok && n > 0 {
			return &deleteOlderThan{n: n}, nil

		} else if ok {
			return nil, ErrInvalidMetadata(MetaDataKeyN, "cannot be negative")
		}

		return nil, ErrWrongMetadataType(MetaDataKeyN, "int")
	}

	return nil, ErrMissingMetadata(MetaDataKeyN)
}

func (*deleteOlderThan) InitializeFor(project *models.Project, repo *models.RepoRecord) {}

func (d *deleteOlderThan) Process(tag *retention.TagRecord) (retention.FilterAction, error) {
	if tag.CreatedAt.Before(time.Now().Add(time.Duration(d.n) * daysAgo)) {
		return retention.FilterActionDelete, nil
	}

	return retention.FilterActionNoDecision, nil
}
