package handler

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
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

func (s *sysInfoAPI) GetSysteminfo(ctx context.Context, params systeminfo.GetSysteminfoParams) middleware.Responder {
	opt := si.Options{}
	sc, ok := security.FromContext(ctx)
	if ok && sc.IsAuthenticated() {
		opt.WithProtectedInfo = true
	}
	data, err := s.ctl.GetInfo(ctx, opt)
	if err != nil {
		return s.SendError(ctx, err)
	}
	return systeminfo.NewGetSysteminfoOK().WithPayload(s.convertInfo(data))
}

func (s *sysInfoAPI) GetSysteminfoGetcert(ctx context.Context, params systeminfo.GetSysteminfoGetcertParams) middleware.Responder {
	f, err := s.ctl.GetCA(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}
	return systeminfo.NewGetSysteminfoGetcertOK().WithContentDisposition("attachment; filename=ca.crt").WithPayload(f)
}

func (s *sysInfoAPI) GetSysteminfoVolumes(ctx context.Context, params systeminfo.GetSysteminfoVolumesParams) middleware.Responder {
	if err := s.RequireSysAdmin(ctx); err != nil {
		return s.SendError(ctx, err)
	}
	c, err := s.ctl.GetCapacity(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}
	return systeminfo.NewGetSysteminfoVolumesOK().WithPayload(&models.SystemInfo{
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
		SelfRegistration: &d.SelfRegistration,
		HarborVersion:    &d.HarborVersion,
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
		res.ExternalURL = &d.Protected.ExtURL
		res.RegistryURL = &d.Protected.RegistryURL
		res.WithChartmuseum = &d.Protected.WithChartMuseum
		res.WithNotary = &d.Protected.WithNotary
		res.ReadOnly = &d.Protected.ReadOnly
		res.RegistryStorageProviderName = &d.Protected.RegistryStorageProviderName
		res.NotificationEnable = &d.Protected.NotificationEnable
	}
	return res

}
