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

package art

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/pkg/errors"
)

// basicManager is the default implementation of artifact manager
type basicManager struct{}

// NewManager creates a new basic manager as the default one.
func NewManager() Manager {
	return &basicManager{}
}

// List artifacts
func (b *basicManager) List(query *q.Query) ([]*models.Artifact, error) {
	aq := &models.ArtifactQuery{}
	makeArtQuery(aq, query)

	l, err := dao.ListArtifacts(aq)
	if err != nil {
		return nil, errors.Wrap(err, "artifact manager: list")
	}

	return l, nil
}

func makeArtQuery(aq *models.ArtifactQuery, query *q.Query) {
	if aq == nil {
		return // do nothing
	}

	if query != nil {
		if len(query.Keywords) > 0 {
			for k, v := range query.Keywords {
				switch k {
				case "project_id":
					aq.PID = v.(int64)
				case "repo":
					aq.Repo = v.(string)
				case "tag":
					aq.Tag = v.(string)
				case "digest":
					aq.Digest = v.(string)
				default:
				}
			}
		}

		if query.PageNumber > 0 && query.PageSize > 0 {
			aq.Page = query.PageNumber
			aq.Size = query.PageSize
		}
	}
}
