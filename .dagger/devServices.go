package main

import (
	"context"
	"dagger/harbor/internal/dagger"
)

func (m *Harbor) NginxService(ctx context.Context) *dagger.Service {
	nginxConfig := m.OnlyDagger.File("./.dagger/config/proxy/nginx.conf")

	nginxSrv := m.BuildImage(ctx, DEV_PLATFORM, "nginx", true).
		// nginxSrv := dag.Container().From("goharbor/nginx-photon:dev").
		WithMountedFile("/etc/nginx/nginx.conf", nginxConfig).
		WithExposedPort(8080).
		// for debug
		WithExposedPort(4001).
		WithoutExposedPort(8443).
		AsService()
	return nginxSrv
}

// not working as expected
func (m *Harbor) PortalService(ctx context.Context) *dagger.Service {
	nginxConfig := m.OnlyDagger.File("./.dagger/config/proxy/nginx.conf")

	portal := m.BuildImage(ctx, DEV_PLATFORM, "portal", true).
		// portal := dag.Container().From("goharbor/harbor-portal:dev").
		WithMountedFile("/etc/nginx/nginx.conf", nginxConfig).
		WithExposedPort(8080).
		WithoutExposedPort(8443).
		AsService()
	return portal
}

func (m *Harbor) JobService(ctx context.Context) *dagger.Service {
	jobSrvConfig := m.OnlyDagger.File("./.dagger/config/jobservice/config.yml")
	envFile := m.OnlyDagger.File("./.dagger/config/jobservice/env")
	run_script := m.OnlyDagger.File("./.dagger/config/run_env.sh")

	jobSrv := m.BuildImage(ctx, DEV_PLATFORM, "jobservice", true).
		WithMountedFile("/etc/jobservice/config.yml", jobSrvConfig).
		WithMountedDirectory("/var/log/jobs", m.OnlyDagger.Directory("./.dagger/config/jobservice")).
		WithMountedFile("/envFile", envFile).
		WithMountedFile("/run_script", run_script).
		WithExec([]string{"chmod", "+x", "/run_script"}).
		WithExposedPort(8080).
		WithEntrypoint([]string{"/run_script", "/jobservice -c /etc/jobservice/config.yml"}).
		AsService()
	return jobSrv
}

func (m *Harbor) CoreService(ctx context.Context) *dagger.Service {
	coreConfig := m.OnlyDagger.File("./.dagger/config/core/app.conf")
	envFile := m.OnlyDagger.File("./.dagger/config/core/env")
	run_script := m.OnlyDagger.File("./.dagger/config/run_debug.sh")
	// run_script := m.OnlyDagger.File("./.dagger/config/run_env.sh")

	core := m.BuildImage(ctx, DEV_PLATFORM, "core", true).
		WithMountedFile("/etc/core/app.conf", coreConfig).
		WithMountedFile("/envFile", envFile).
		WithMountedFile("/run_script", run_script).
		// why alpine instead of golang. because we get the below error
		// [INFO] [/src/common/dao/base.go:72]: Register database completed
		// [FATAL] [/src/core/main.go:203]: failed to migrate the database, error: open .: no such file or directory
		WithExposedPort(8080, dagger.ContainerWithExposedPortOpts{ExperimentalSkipHealthcheck: true}).
		// WithServiceBinding("redis", m.RedisService(ctx)).
		// WithServiceBinding("postgresql", m.PostgresService(ctx)).
		// WithServiceBinding("registry", m.RegistryService(ctx)).
		// WithServiceBinding("registryctl", m.RegistryCtlService(ctx)).
		// WithExposedPort(80).
		WithExposedPort(4001, dagger.ContainerWithExposedPortOpts{ExperimentalSkipHealthcheck: true}).
		// WithEntrypoint([]string{"/root/go/bin/dlv", "--headless=true", "--listen=localhost:4001", "--accept-multiclient", "--log-output=debugger,debuglineerr,gdbwire,lldbout,rpc", "--log=true", "--api-config=2", "exec", "/core"}).
		// WithEntrypoint([]string{"/run_script", "/root/go/bin/dlv --headless=true --listen=localhost:4001 --accept-multiclient --log-output=debugger,debuglineerr,gdbwire,lldbout,rpc --log=true --api-config=2 exec /core"}).
		WithEntrypoint([]string{"/run_script", "/core", "4001"}).
		AsService(dagger.ContainerAsServiceOpts{InsecureRootCapabilities: true})

	return core
}

func (m *Harbor) RegistryCtlService(ctx context.Context) *dagger.Service {
	regConfigDir := m.OnlyDagger.Directory("./.dagger/config/registry")
	regCtlConfig := m.OnlyDagger.File("./.dagger/config/registryctl/config.yml")
	envFile := m.OnlyDagger.File("./.dagger/config/jobservice/env")
	run_script := m.OnlyDagger.File("./.dagger/config/run_env.sh")

	regCtl := m.BuildImage(ctx, DEV_PLATFORM, "registryctl", true).
		WithMountedDirectory("/etc/registry", regConfigDir).
		WithMountedFile("/etc/registryctl/config.yml", regCtlConfig).
		WithMountedFile("/envFile", envFile).
		WithMountedFile("/run_script", run_script).
		WithEntrypoint([]string{"/run_script", "/registryctl -c /etc/registryctl/config.yml"}).
		AsService()

	return regCtl
}

func (m *Harbor) PostgresService(ctx context.Context) *dagger.Service {
	version := m.GetVersion(ctx)
	postgres := dag.Container().From("goharbor/harbor-db:"+version).
		WithExposedPort(5432).
		WithEnvVariable("POSTGRES_PASSWORD", "root123").
		AsService()
	return postgres
}

func (m *Harbor) RedisService(ctx context.Context) *dagger.Service {
	version := m.GetVersion(ctx)
	return dag.Container().
		From("goharbor/redis-photon:" + version).
		WithExposedPort(6379).
		AsService()
}

func (m *Harbor) RegistryService(ctx context.Context) *dagger.Service {
	regConfigDir := m.OnlyDagger.Directory("./.dagger/config/registry")

	// 5001 is can be used to debug according to config
	reg := m.BuildImage(ctx, DEV_PLATFORM, "registry", true).
		WithMountedDirectory("/etc/registry", regConfigDir).
		WithExposedPort(5000).
		WithoutExposedPort(5001).
		WithoutExposedPort(5443).
		AsService()
	return reg
}
