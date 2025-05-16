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

package handler

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	si "github.com/goharbor/harbor/src/controller/systeminfo"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/systeminfo"
)

type sysInfoAPI struct {
	BaseAPI
	ctl si.Controller
}

func newSystemInfoAPI() *sysInfoAPI {
	return &sysInfoAPI{
		ctl: si.Ctl,
	}
}

func (s *sysInfoAPI) GetSystemInfo(ctx context.Context, _ systeminfo.GetSystemInfoParams) middleware.Responder {
	opt := si.Options{}
	sc, ok := security.FromContext(ctx)
	if ok && sc.IsAuthenticated() {
		opt.WithProtectedInfo = true
	}
	data, err := s.ctl.GetInfo(ctx, opt)
	if err != nil {
		return s.SendError(ctx, err)
	}
	return systeminfo.NewGetSystemInfoOK().WithPayload(s.convertInfo(data))
}

func (s *sysInfoAPI) GetCert(ctx context.Context, _ systeminfo.GetCertParams) middleware.Responder {
	f, err := s.ctl.GetCA(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}
	return systeminfo.NewGetCertOK().WithContentDisposition("attachment; filename=ca.crt").WithPayload(f)
}

func (s *sysInfoAPI) GetVolumes(ctx context.Context, _ systeminfo.GetVolumesParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceSystemVolumes); err != nil {
		return s.SendError(ctx, err)
	}
	c, err := s.ctl.GetCapacity(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}
	return systeminfo.NewGetVolumesOK().WithPayload(&models.SystemInfo{
		Storage: []*models.Storage{
			{
				Free:  c.Free,
				Total: c.Total,
			},
		},
	})
}

func (s *sysInfoAPI) convertInfo(d *si.Data) *models.GeneralInfo {
	if d == nil {
		return nil
	}
	res := &models.GeneralInfo{
		AuthMode:         &d.AuthMode,
		PrimaryAuthMode:  &d.PrimaryAuthMode,
		SelfRegistration: &d.SelfRegistration,
		BannerMessage:    &d.BannerMessage,
		OIDCProviderName: &d.OIDCProviderName,
	}
	if d.AuthProxySettings != nil {
		res.AuthproxySettings = &models.AuthproxySetting{
			Endpoint:            d.AuthProxySettings.Endpoint,
			TokenreivewEndpoint: d.AuthProxySettings.TokenReviewEndpoint,
			ServerCertificate:   d.AuthProxySettings.ServerCertificate,
			VerifyCert:          d.AuthProxySettings.VerifyCert,
			SkipSearch:          d.AuthProxySettings.SkipSearch,
		}
	}

	if d.Protected != nil {
		res.HasCaRoot = &d.Protected.HasCARoot
		res.ProjectCreationRestriction = &d.Protected.ProjectCreationRestrict
		res.HarborVersion = &d.Protected.HarborVersion
		res.ExternalURL = &d.Protected.ExtURL
		res.RegistryURL = &d.Protected.RegistryURL
		res.ReadOnly = &d.Protected.ReadOnly
		res.RegistryStorageProviderName = &d.Protected.RegistryStorageProviderName
		res.NotificationEnable = &d.Protected.NotificationEnable
		currentTime := strfmt.DateTime(d.Protected.CurrentTime)
		res.CurrentTime = &currentTime
	}
	return res
}
