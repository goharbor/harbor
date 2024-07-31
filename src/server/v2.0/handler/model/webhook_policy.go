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
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// WebhookPolicy ...
type WebhookPolicy struct {
	*model.Policy
}

// ToSwagger ...
func (n *WebhookPolicy) ToSwagger() *models.WebhookPolicy {
	return &models.WebhookPolicy{
		ID:           n.ID,
		CreationTime: strfmt.DateTime(n.CreationTime),
		UpdateTime:   strfmt.DateTime(n.UpdateTime),
		Creator:      n.Creator,
		Description:  n.Description,
		Enabled:      n.Enabled,
		EventTypes:   n.EventTypes,
		Name:         n.Name,
		ProjectID:    n.ProjectID,
		Targets:      n.ToTargets(),
	}
}

// ToTargets ...
func (n *WebhookPolicy) ToTargets() []*models.WebhookTargetObject {
	var results []*models.WebhookTargetObject
	for _, t := range n.Targets {
		results = append(results, &models.WebhookTargetObject{
			Type:           t.Type,
			Address:        t.Address,
			AuthHeader:     t.AuthHeader,
			SkipCertVerify: t.SkipCertVerify,
			PayloadFormat:  models.PayloadFormatType(t.PayloadFormat),
		})
	}
	return results
}

// NewWebhookPolicy ...
func NewWebhookPolicy(p *model.Policy) *WebhookPolicy {
	return &WebhookPolicy{
		Policy: p,
	}
}
