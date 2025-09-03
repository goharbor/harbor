package main

import (
	"context"
	"dagger/harbor/internal/dagger"
)

// Run Harbor inside Dagger
func (m *Harbor) DevServer(ctx context.Context) (*dagger.Service, error) {
	var err error

	postgresSrv := m.PostgresService(ctx)
	redisSrv := m.RedisService(ctx)
	regSrv := m.RegistryService(ctx)
	regCtlSrv := m.RegistryCtlService(ctx)
	coreSrv := m.CoreService(ctx)
	jobSrv := m.JobService(ctx)
	portalSrv := m.PortalService(ctx)
	nginxSrv := m.NginxService(ctx)

	_, err = postgresSrv.WithHostname("postgresql").Start(ctx)
	if err != nil {
		return nil, err
	}
	_, err = redisSrv.WithHostname("redis").Start(ctx)
	if err != nil {
		return nil, err
	}
	_, err = regSrv.WithHostname("registry").Start(ctx)
	if err != nil {
		return nil, err
	}
	_, err = regCtlSrv.WithHostname("registryctl").Start(ctx)
	if err != nil {
		return nil, err
	}
	_, err = coreSrv.WithHostname("core").Start(ctx)
	if err != nil {
		return nil, err
	}
	_, err = jobSrv.WithHostname("jobservice").Start(ctx)
	if err != nil {
		return nil, err
	}
	_, err = portalSrv.WithHostname("portal").Start(ctx)
	if err != nil {
		return nil, err
	}
	proxy, err := nginxSrv.WithHostname("proxy").Start(ctx)
	if err != nil {
		return nil, err
	}

	return proxy, nil
}
