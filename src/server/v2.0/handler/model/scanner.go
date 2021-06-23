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
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// ScannerRegistration ...
type ScannerRegistration struct {
	*scanner.Registration
}

// ToSwagger ...
func (s *ScannerRegistration) ToSwagger(ctx context.Context) *models.ScannerRegistration {
	if s.Registration == nil {
		return nil
	}

	return &models.ScannerRegistration{
		UUID:             s.UUID,
		Name:             s.Name,
		URL:              s.URL,
		Description:      s.Description,
		Auth:             s.Auth,
		AccessCredential: s.AccessCredential,
		SkipCertVerify:   &s.SkipCertVerify,
		UseInternalAddr:  &s.UseInternalAddr,
		IsDefault:        &s.IsDefault,
		Disabled:         &s.Disabled,
		CreateTime:       strfmt.DateTime(s.CreateTime),
		UpdateTime:       strfmt.DateTime(s.UpdateTime),
		Adapter:          s.Adapter,
		Vendor:           s.Vendor,
		Version:          s.Version,
		Health:           s.Health,
	}
}

// NewScannerRegistration ...
func NewScannerRegistration(scanner *scanner.Registration) *ScannerRegistration {
	return &ScannerRegistration{Registration: scanner}
}

// ScannerMetadata ...
type ScannerMetadata struct {
	*v1.ScannerAdapterMetadata
}

// ToSwagger ...
func (s *ScannerMetadata) ToSwagger(ctx context.Context) *models.ScannerAdapterMetadata {
	if s.ScannerAdapterMetadata == nil {
		return nil
	}

	var capabilities []*models.ScannerCapability
	for _, c := range s.Capabilities {
		capabilities = append(capabilities, &models.ScannerCapability{
			ConsumesMimeTypes: c.ConsumesMimeTypes,
			ProducesMimeTypes: c.ProducesMimeTypes,
		})
	}
	return &models.ScannerAdapterMetadata{
		Scanner:      (*models.Scanner)(s.Scanner),
		Properties:   s.Properties,
		Capabilities: capabilities,
	}
}

// NewScannerMetadata ...
func NewScannerMetadata(md *v1.ScannerAdapterMetadata) *ScannerMetadata {
	return &ScannerMetadata{ScannerAdapterMetadata: md}
}
