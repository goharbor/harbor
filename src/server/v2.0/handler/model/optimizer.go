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

	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
	v1 "github.com/goharbor/harbor/src/pkg/optimizer/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// OptimizerRegistration ...
type OptimizerRegistration struct {
	*dao.Registration
}

// ToSwagger ...
func (s *OptimizerRegistration) ToSwagger(_ context.Context) *models.OptimizerRegistration {
	if s.Registration == nil {
		return nil
	}

	return &models.OptimizerRegistration{
		UUID:            s.UUID,
		Name:            s.Name,
		URL:             strfmt.URI(s.URL),
		Description:     s.Description,
		Auth:            s.Auth,
		SkipCertVerify:  &s.SkipCertVerify,
		UseInternalAddr: &s.UseInternalAddr,
		IsDefault:       &s.IsDefault,
		Disabled:        &s.Disabled,
		CreateTime:      strfmt.DateTime(s.CreateTime),
		UpdateTime:      strfmt.DateTime(s.UpdateTime),
		Adapter:         s.Adapter,
		Vendor:          s.Vendor,
		Version:         s.Version,
		Health:          s.Health,
	}
}

// NewOptimizerRegistration ...
func NewOptimizerRegistration(registration *dao.Registration) *OptimizerRegistration {
	return &OptimizerRegistration{Registration: registration}
}

// OptimizerMetadata ...
type OptimizerMetadata struct {
	*v1.OptimizerAdapterMetadata
}

// ToSwagger ...
func (s *OptimizerMetadata) ToSwagger(_ context.Context) *models.OptimizerAdapterMetadata {
	if s.OptimizerAdapterMetadata == nil {
		return nil
	}

	var capabilities []*models.OptimizerCapability
	for _, c := range s.Capabilities {
		capabilities = append(capabilities, &models.OptimizerCapability{
			ConsumesMimeTypes: c.ConsumesMimeTypes,
			ProducesMimeTypes: c.ProducesMimeTypes,
		})
	}
	return &models.OptimizerAdapterMetadata{
		Optimizer:    (*models.Optimizer)(s.Optimizer),
		Properties:   s.Properties,
		Capabilities: capabilities,
	}
}

// NewOptimizerMetadata ...
func NewOptimizerMetadata(md *v1.OptimizerAdapterMetadata) *OptimizerMetadata {
	return &OptimizerMetadata{OptimizerAdapterMetadata: md}
}
