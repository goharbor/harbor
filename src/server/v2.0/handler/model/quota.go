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

package model

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Quota model
type Quota struct {
	*quota.Quota
}

// ToSwagger converts the quota to the swagger model
func (q *Quota) ToSwagger(ctx context.Context) *models.Quota {
	if q.Quota == nil {
		return nil
	}

	hard, err := q.GetHard()
	if err != nil {
		fields := log.Fields{"quota_id": q.ID, "error": err}
		log.G(ctx).WithFields(fields).Warningf("failed to get hard from quota")

		hard = types.ResourceList{}
	}

	used, err := q.GetUsed()
	if err != nil {
		fields := log.Fields{"quota_id": q.ID, "error": err}
		log.G(ctx).WithFields(fields).Warningf("failed to get used from quota")

		used = types.ResourceList{}
	}

	return &models.Quota{
		ID:           q.ID,
		Ref:          q.Ref,
		Hard:         NewResourceList(hard).ToSwagger(),
		Used:         NewResourceList(used).ToSwagger(),
		CreationTime: strfmt.DateTime(q.CreationTime),
		UpdateTime:   strfmt.DateTime(q.UpdateTime),
	}
}

// NewQuota new quota instance
func NewQuota(quota *quota.Quota) *Quota {
	return &Quota{Quota: quota}
}
