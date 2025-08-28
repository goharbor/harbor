package main

import (
	"context"
	"dagger/harbor/internal/dagger"
	"fmt"
	"log"
	"strings"
)

type (
	Package  string
	Platform string
)

var (
	targetPlatforms = []Platform{"linux/amd64", "linux/arm64"}
	packages        = []Package{"core", "jobservice", "registryctl", "portal", "registry", "nginx", "cmd/exporter", "trivy-adapter"}
)

type BuildMetadata struct {
	Package    Package
	BinaryPath string
	Container  *dagger.Container
	Platform   Platform
}

func New(
	// +optional
	// +defaultPath="./"
	// +ignore=["bin", "node_modules"]
	source *dagger.Directory,
	// +optional
	// +defaultPath="./"
	// +ignore=[".dagger", "node_modules", ".github", "contrib", "docs", "icons", "tests", "make", "bin", "*.md"]
	filteredSrc *dagger.Directory,
	// +optional
	// +defaultPath="./"
	// +ignore=[".dagger", "node_modules", ".github", "contrib", "docs", "icons", "tests", "make", "bin", "!src/portal/**", "!LICENSE"]
	portalSrc *dagger.Directory,
	// +optional
	// +defaultPath="./"
	// +ignore=["*", "!.dagger"]
	onlyDagger *dagger.Directory,
) *Harbor {
	return &Harbor{Source: source, FilteredSrc: filteredSrc, PortalSrc: portalSrc, OnlyDagger: onlyDagger}
}

type Harbor struct {
	Source      *dagger.Directory
	FilteredSrc *dagger.Directory
	PortalSrc   *dagger.Directory
	OnlyDagger  *dagger.Directory
}

// build, publish and sign image
func (m *Harbor) PublishAndSignImage(
	ctx context.Context,
	registry, registryUsername, projectName string,
	pkg Package,
	debugbin bool,
	imageTags []string,
	registryPassword *dagger.Secret,
	// +optional
	sigstoreIdToken *dagger.Secret,
) (string, error) {
	imageAddrs := m.PublishImage(ctx, registry, registryUsername, projectName, imageTags, debugbin, pkg, registryPassword)
	_, err := m.Sign(
		ctx,
		sigstoreIdToken,
		registryUsername,
		registryPassword,
		imageAddrs[0],
	)
	if err != nil {
		return "", fmt.Errorf("failed to sign image: %w", err)
	}

	fmt.Printf("Signed image: %s\n", imageAddrs)
	return imageAddrs[0], nil
}

// build, publish and sign all images
func (m *Harbor) PublishAndSignAllImages(
	ctx context.Context,
	registry string,
	registryUsername string,
	projectName string,
	registryPassword *dagger.Secret,
	debugbin bool,
	imageTags []string,
	// +optional
	sigstoreIdToken *dagger.Secret,
) (string, error) {
	imageAddrs := m.PublishAllImages(ctx, registry, registryUsername, projectName, imageTags, debugbin, registryPassword)
	_, err := m.Sign(
		ctx,
		sigstoreIdToken,
		registryUsername,
		registryPassword,
		imageAddrs[0],
	)
	if err != nil {
		return "", fmt.Errorf("failed to sign image: %w", err)
	}

	fmt.Printf("Signed image: %s\n", imageAddrs)
	return imageAddrs[0], nil
}

// Sign signs a container image using Cosign, works also with GitHub Actions
func (m *Harbor) Sign(ctx context.Context,
	// +optional
	sigstoreIdToken *dagger.Secret,
	registryUsername string,
	registryPassword *dagger.Secret,
	imageAddr string,
) (string, error) {
	registryPasswordPlain, _ := registryPassword.Plaintext(ctx)

	cosing_ctr := dag.Container().From("cgr.dev/chainguard/cosign")

	// If githubToken is provided, use it to sign the image. (GitHub Actions) use case
	if sigstoreIdToken != nil {
		fmt.Printf("Setting the ENV Vars SIGSTORE_ID_TOKEN to sign with Token")
		cosing_ctr = cosing_ctr.WithSecretVariable("SIGSTORE_ID_TOKEN", sigstoreIdToken)
	}

	return cosing_ctr.WithSecretVariable("REGISTRY_PASSWORD", registryPassword).
		WithExec([]string{"cosign", "env"}).
		WithExec([]string{
			"cosign", "sign", "--yes", "--recursive",
			"--registry-username", registryUsername,
			"--registry-password", registryPasswordPlain,
			imageAddr,
			"--timeout", "1m",
		}).Stdout(ctx)
}

// Publishes All Images and variants
func (m *Harbor) PublishAllImages(
	ctx context.Context,
	registry, registryUsername, projectName string,
	imageTags []string,
	debugbin bool,
	registryPassword *dagger.Secret,
) []string {
	fmt.Printf("provided tags: %s\n", imageTags)

	allImages := m.buildAllImages(ctx, debugbin)
	platformVariantsContainer := make(map[Package][]*dagger.Container)
	for _, meta := range allImages {
		platformVariantsContainer[meta.Package] = append(platformVariantsContainer[meta.Package], meta.Container)
	}

	var imageAddresses []string
	var imgAddress string
	var err error
	for pkg, imgs := range platformVariantsContainer {
		for _, imageTag := range imageTags {
			container := dag.Container().WithRegistryAuth(registry, registryUsername, registryPassword)
			if pkg == "cmd/exporter" {
				imgAddress, err = container.Publish(ctx,
					fmt.Sprintf("%s/%s/%s:%s", registry, projectName, "harbor-exporter", imageTag),
					dagger.ContainerPublishOpts{PlatformVariants: imgs},
				)
			} else if pkg == "trivy-adapter" {
				imgAddress, err = container.Publish(ctx,
					fmt.Sprintf("%s/%s/%s:%s", registry, projectName, "trivy-adapter", imageTag),
					dagger.ContainerPublishOpts{PlatformVariants: imgs},
				)
			} else {
				imgAddress, err = container.Publish(ctx,
					fmt.Sprintf("%s/%s/%s:%s", registry, projectName, "harbor-"+pkg, imageTag),
					dagger.ContainerPublishOpts{PlatformVariants: imgs},
				)
			}
			if err != nil {
				fmt.Printf("Failed to publish image: %s/%s/%s:%s\n", registry, projectName, pkg, imageTag)
				fmt.Printf("Error: %s\n", err)
				continue
			}
			imageAddresses = append(imageAddresses, imgAddress)
			fmt.Printf("Published image: %s\n", imgAddress)
		}
	}
	return imageAddresses
}

// publishes the specific image with the given tag
func (m *Harbor) PublishImage(
	ctx context.Context,
	registry, registryUsername, projectName string,
	imageTags []string,
	debugbin bool,
	pkg Package,
	registryPassword *dagger.Secret,
) []string {
	var (
		imageAddresses []string
		images         []*dagger.Container
	)

	fmt.Printf("provided tags: %s\n", imageTags)

	for _, platform := range targetPlatforms {
		BuildImage := m.BuildImage(ctx, platform, pkg, debugbin)
		images = append(images, BuildImage)
	}

	platformVariantsContainer := make(map[Package][]*dagger.Container)
	for _, image := range images {
		platformVariantsContainer[pkg] = append(platformVariantsContainer[pkg], image)
	}

	var imgAddress string
	var err error
	for pkg, imgs := range platformVariantsContainer {
		for _, imageTag := range imageTags {
			container := dag.Container().WithRegistryAuth(registry, registryUsername, registryPassword)
			if pkg == "cmd/exporter" {
				imgAddress, err = container.Publish(ctx,
					fmt.Sprintf("%s/%s/%s:%s", registry, projectName, "harbor-exporter", imageTag),
					dagger.ContainerPublishOpts{PlatformVariants: imgs},
				)
			} else if pkg == "trivy-adapter" {
				imgAddress, err = container.Publish(ctx,
					fmt.Sprintf("%s/%s/%s:%s", registry, projectName, "trivy-adapter", imageTag),
					dagger.ContainerPublishOpts{PlatformVariants: imgs},
				)
			} else {
				imgAddress, err = container.Publish(ctx,
					fmt.Sprintf("%s/%s/%s:%s", registry, projectName, "harbor-"+pkg, imageTag),
					dagger.ContainerPublishOpts{PlatformVariants: imgs},
				)
			}
			if err != nil {
				fmt.Printf("Failed to publish image: %s/%s/%s:%s\n", registry, projectName, pkg, imageTag)
				fmt.Printf("Error: %s\n", err)
				continue
			}
			imageAddresses = append(imageAddresses, imgAddress)
			fmt.Printf("Published image: %s\n", imgAddress)
		}
	}

	return imageAddresses
}

// export all images as Tarball
func (m *Harbor) ExportAllImages(ctx context.Context, debugbin bool) *dagger.Directory {
	metdata := m.buildAllImages(ctx, debugbin)
	artifacts := dag.Directory()
	for _, meta := range metdata {
		artifacts = artifacts.WithFile(fmt.Sprintf("containers/%s/%s.tgz", meta.Platform, meta.Package), meta.Container.AsTarball())
	}
	return artifacts
}

// build all images
func (m *Harbor) BuildAllImages(ctx context.Context, debugbin bool) []*dagger.Container {
	metdata := m.buildAllImages(ctx, debugbin)
	images := make([]*dagger.Container, len(metdata))
	for i, meta := range metdata {
		images[i] = meta.Container
	}
	return images
}

func (m *Harbor) buildAllImages(ctx context.Context, debugbin bool) []*BuildMetadata {
	var (
		buildMetadata []*BuildMetadata // final result
	)

	for _, platform := range targetPlatforms {
		for _, pkg := range packages {
			platform := platform
			pkg := pkg

			img := m.BuildImage(ctx, platform, pkg, debugbin)

			metadata := &BuildMetadata{
				Package:    pkg,
				BinaryPath: fmt.Sprintf("bin/%s/%s", platform, pkg),
				Container:  img,
				Platform:   platform,
			}
			buildMetadata = append(buildMetadata, metadata)
		}
	}

	return buildMetadata
}

// build single specified images
func (m *Harbor) BuildImage(ctx context.Context, platform Platform, pkg Package, debugbin bool) *dagger.Container {
	buildMtd := m.buildImage(ctx, platform, pkg, debugbin)
	if pkg == "core" {
		// the only thing missing here is the healthcheck
		// we can add those by updating the docker compose since dagger currently doesn't support healthchecks
		// issue: https://github.com/dagger/dagger/issues/9515
		buildMtd.Container = buildMtd.Container.WithDirectory("/migrations", m.Source.Directory("make/migrations")).
			WithDirectory("/icons", m.Source.Directory("icons")).
			WithDirectory("/views", m.Source.Directory("src/core/views")).
			WithWorkdir("/")
	}
	if pkg == "registryctl" {
		regBinary := m.registryBuilder(ctx, platform)
		buildMtd.Container = buildMtd.Container.WithFile("/usr/bin/registry_DO_NOT_USE_GC", regBinary).
			WithExposedPort(8080)
	}

	return buildMtd.Container
}

// deprecated: internal function to build registry
func (m *Harbor) registryBuilder(ctx context.Context, platform Platform) *dagger.File {
	registrySrc := dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(string(platform))}).
		From("golang:"+GO_VERSION+"-alpine").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod-"+GO_VERSION)).
		WithMountedCache("/go/build-cache", dag.CacheVolume("go-build-"+GO_VERSION)).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithEnvVariable("GOCACHE", "/go/build-cache").
		WithEnvVariable("DISTRIBUTION_DIR", "/go/src/github.com/docker/distribution").
		WithEnvVariable("BUILDTAGS", "include_oss include_gcs").
		WithEnvVariable("GO111MODULE", "auto").
		WithEnvVariable("CGO_ENABLED", "0").
		WithWorkdir("/go/src/github.com/docker").
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithExec([]string{"git", "clone", "-b", REGISTRY_SRC_TAG, DISTRIBUTION_SRC}).
		WithWorkdir("distribution").
		// fix for CVE-2025-22872 https://avd.aquasec.com/nvd/2025/cve-2025-22872/
		WithExec([]string{"go", "mod", "edit", "-require", "golang.org/x/net@v0.38.0"}).
		// update & clean
		WithExec([]string{"go", "mod", "tidy", "-e"}).
		WithExec([]string{"go", "mod", "vendor"}).
		// comment out when using v3
		// WithFile("/redis.patch", m.OnlyDagger.File("./.dagger/registry/redis.patch")).
		// WithExec([]string{"git", "apply", "/redis.patch"}).
		WithExec([]string{"echo", "build the registry binary"})

	// created based on distribution's dockerfile
	// https://github.com/distribution/distribution/blob/main/Dockerfile
	// 'version' stage: Generate versioning info and linker flags
	// This container will generate the .ldflags file
	versioner := registrySrc.WithExec([]string{
		"sh", "-c",
		`VERSION=$(git describe --match 'v[0-9]*' --dirty='.m' --always --tags) && \
		 REVISION=$(git rev-parse HEAD) && \
		 PKG=github.com/distribution/distribution/v3 && \
		 echo "-X ${PKG}/version.version=${VERSION#v} -X ${PKG}/version.revision=${REVISION} -X ${PKG}/version.mainpkg=${PKG}" > /tmp/.ldflags`,
	})

	ldflagsFile := versioner.File("/tmp/.ldflags")

	// build stage
	builder := registrySrc.
		// Mount the ldflags file from the 'versioner' container
		WithFile("/tmp/.ldflags", ldflagsFile).
		WithExec([]string{
			"sh", "-c",
			`CGO_ENABLED=0 go build -trimpath -ldflags "$(cat /tmp/.ldflags) -s -w" -o /go/bin/registry ./cmd/registry`,
		}).
		WithExec([]string{"/go/bin/registry", "--version"})

	// 5. Extract the final binary
	registryBinary := builder.File("/go/bin/registry")

	return registryBinary
}

func (m *Harbor) buildImage(ctx context.Context, platform Platform, pkg Package, debugbin bool) *BuildMetadata {
	var (
		buildMtd *BuildMetadata
		img      *dagger.Container
	)

	if pkg == "trivy-adapter" {
		img = m.buildTrivyAdapter(ctx, platform)
		buildMtd = &BuildMetadata{
			Package:    pkg,
			BinaryPath: "nil",
			Container:  img,
			Platform:   platform,
		}
	} else if pkg == "portal" {
		img = m.buildPortal(ctx, platform)
		buildMtd = &BuildMetadata{
			Package:    pkg,
			BinaryPath: "nil",
			Container:  img,
			Platform:   platform,
		}
	} else if pkg == "registry" {
		img = m.buildRegistry(ctx, platform)
		buildMtd = &BuildMetadata{
			Package:    pkg,
			BinaryPath: "nil",
			Container:  img,
			Platform:   platform,
		}
	} else if pkg == "nginx" {
		img = m.buildNginx(ctx, platform)
		buildMtd = &BuildMetadata{
			Package:    pkg,
			BinaryPath: "nil",
			Container:  img,
			Platform:   platform,
		}
	} else {
		buildMtd = m.buildBinary(ctx, platform, pkg, debugbin)
		img = dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(string(platform))}).
			WithDirectory("/etc/ssl/certs", m.getCaCerts(ctx)).
			WithFile("/"+string(pkg), buildMtd.Container.File(buildMtd.BinaryPath))

		// Set entrypoint based on package
		entrypoint := []string{"/" + string(pkg)}
		if pkg == "jobservice" {
			entrypoint = append(entrypoint, "-c", "/etc/jobservice/config.yml")
		} else if pkg == "registryctl" {
			entrypoint = append(entrypoint, "-c", "/etc/registryctl/config.yml")
		}

		// handle for debug
		if debugbin && string(platform) == "linux/amd64" {
			fmt.Println("entrypoint before:", entrypoint)
			entrycmd := entrypoint[0]
			entryargs := strings.Join(entrypoint[1:], " ")
			debug_entrypoint := fmt.Sprintf("%s -- %s", entrycmd, entryargs)

			fmt.Println("entrypoint for debug:", debug_entrypoint)

			img = dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(string(platform))}).
				From("golang:"+GO_VERSION+"-alpine").
				WithDirectory("/etc/ssl/certs", m.getCaCerts(ctx)).
				WithExec([]string{"go", "install", "github.com/go-delve/delve/cmd/dlv@" + DELVE_VERSION}).
				WithExposedPort(8080).
				// should use script since executing with config would result in an error
				WithExposedPort(4001, dagger.ContainerWithExposedPortOpts{ExperimentalSkipHealthcheck: true}).
				WithFile("/"+string(pkg), buildMtd.Container.File(buildMtd.BinaryPath)).
				WithFile("/entrypoint.sh", m.OnlyDagger.File("./.dagger/config/debug_entrypoint.sh")).
				WithExec([]string{"chmod", "+x", "/entrypoint.sh"}).
				WithEntrypoint([]string{"/entrypoint.sh", debug_entrypoint})

			if pkg == "core" {
				img = img.WithEntrypoint([]string{"/entrypoint.sh", entrycmd})
			}

		} else {
			img = img.WithEntrypoint(entrypoint)
		}
	}

	buildMtd.Container = img
	return buildMtd
}

// build all binaries and return directory containing all binaries
func (m *Harbor) BuildAllBinaries(ctx context.Context, debugbin bool) *dagger.Directory {
	output := dag.Directory()
	builds := m.buildAllBinaries(ctx, debugbin)
	for _, build := range builds {
		output = output.WithFile(build.BinaryPath, build.Container.File(build.BinaryPath))
	}
	return output
}

func (m *Harbor) buildAllBinaries(ctx context.Context, debugbin bool) []*BuildMetadata {
	var buildContainers []*BuildMetadata
	for _, platform := range targetPlatforms {
		for _, pkg := range packages {
			buildContainer := m.buildBinary(ctx, platform, pkg, debugbin)
			buildContainers = append(buildContainers, buildContainer)
		}
	}
	return buildContainers
}

// builds binary for the specified package
func (m *Harbor) BuildBinary(ctx context.Context, platform Platform, pkg Package, debugbin bool) *dagger.File {
	build := m.buildBinary(ctx, platform, pkg, debugbin)
	return build.Container.File(build.BinaryPath)
}

func (m *Harbor) buildBinary(ctx context.Context, platform Platform, pkg Package, debugbin bool) *BuildMetadata {
	var (
		srcWithSwagger *dagger.Directory
		ldflags        string
		gcflags        string
	)

	ldflags = "-extldflags=-static -s -w"
	goflags := "-buildvcs=false"

	if debugbin {
		gcflags = "all=-N -l"
		ldflags = ""
	}

	os, arch, err := parsePlatform(string(platform))
	if err != nil {
		log.Fatalf("Error parsing platform: %v", err)
	}

	if pkg == "core" {
		m.lintAPIs(ctx).Sync(ctx)
		srcWithSwagger = m.genAPIs(ctx)
		m.FilteredSrc = m.FilteredSrc.WithDirectory("./src/server/v2.0", srcWithSwagger)

		gitCommit := m.fetchGitCommit(ctx)
		version := m.GetVersion(ctx)

		// srcWithSwagger = m.genAPIs(ctx)
		// m.Source = srcWithSwagger
		ldflags = fmt.Sprintf(`-X github.com/goharbor/harbor/src/pkg/version.GitCommit=%s
                    -X github.com/goharbor/harbor/src/pkg/version.ReleaseVersion=%s
    `, gitCommit, version)
	}

	outputPath := fmt.Sprintf("bin/%s/%s", platform, pkg)
	src := fmt.Sprintf("%s/main.go", pkg)
	builder := dag.Container().
		From("golang:"+GO_VERSION+"-alpine").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod-"+GO_VERSION)).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", dag.CacheVolume("go-build-"+GO_VERSION)).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		// update for better caching
		WithMountedDirectory("/src", m.FilteredSrc).
		WithWorkdir("/src/src").
		WithEnvVariable("GOOS", os).
		WithEnvVariable("GOARCH", arch).
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", goflags, "-gcflags=" + gcflags, "-o", outputPath, "-ldflags", ldflags, src})

	return &BuildMetadata{
		Package:    pkg,
		BinaryPath: outputPath,
		Container:  builder,
		Platform:   platform,
	}
}

// internal function to build Nginx
func (m *Harbor) buildNginx(ctx context.Context, platform Platform) *dagger.Container {
	fmt.Println("🛠️  Building Harbor Nginx...")

	return dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(string(platform))}).
		From("nginx:alpine").
		WithDirectory("/etc/ssl/certs", m.getCaCerts(ctx)).
		WithExposedPort(8080).
		WithEntrypoint([]string{"nginx", "-g", "daemon off;"})
}

func (m *Harbor) buildRegistry(ctx context.Context, platform Platform) *dagger.Container {
	fmt.Println("🛠️  Building Harbor Registry...")

	regBinary := m.registryBuilder(ctx, platform)
	// regBinary := m.getRegistry(ctx, platform)
	return dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(string(platform))}).
		// WithExec([]string{"apk", "add", "--no-cache", "libc6-compat"}).
		WithWorkdir("/").
		WithDirectory("/etc/ssl/certs", m.getCaCerts(ctx)).
		WithFile("/usr/bin/registry_DO_NOT_USE_GC", regBinary).
		WithExec([]string{"/usr/bin/registry_DO_NOT_USE_GC", "--version"}).
		// specifically set this for distribution v3
		WithEnvVariable("OTEL_TRACES_EXPORTER", "none").
		WithExposedPort(5000).
		WithExposedPort(5443).
		WithEntrypoint([]string{"/usr/bin/registry_DO_NOT_USE_GC", "serve", "/etc/registry/config.yml"})
}

// internal function to build Trivy Adapter
func (m *Harbor) buildTrivyAdapter(ctx context.Context, platform Platform) *dagger.Container {
	fmt.Println("🛠️  Building Trivy Adapter...")

	trivyBinDir := dag.Container().
		From("golang:"+GO_VERSION).
		WithWorkdir("/go/src/github.com/goharbor/").
		WithExec([]string{"git", "clone", "-b", TRIVYADAPTERVERSION, "https://github.com/goharbor/harbor-scanner-trivy.git"}).
		WithWorkdir("harbor-scanner-trivy").
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", dag.CacheVolume("go-build-"+GO_VERSION)).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		WithEnvVariable("DISTRIBUTION_DIR", "/go/src/github.com/docker/distribution").
		WithEnvVariable("BUILDTAGS", "include_oss include_gcs").
		WithEnvVariable("GO111MODULE", "auto").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", "./binary/scanner-trivy", "cmd/scanner-trivy/main.go"}).
		WithExec([]string{"wget", "-O", "trivyDownload", TRIVY_DOWNLOAD_URL}).
		WithExec([]string{"tar", "-zxv", "-f", "trivyDownload"}).
		WithExec([]string{"cp", "trivy", "./binary/trivy"}).
		Directory("binary")

	trivyAdapter := trivyBinDir.File("./trivy")
	trivyScanner := trivyBinDir.File("./scanner-trivy")

	return dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(string(platform))}).
		From("aquasec/trivy:"+TRIVY_VERSION_NO_PREFIX).
		WithDirectory("/etc/ssl/certs", m.getCaCerts(ctx)).
		WithFile("/home/scanner/bin/scanner-trivy", trivyScanner).
		WithFile("/usr/local/bin/trivy", trivyAdapter).
		// ENV TRIVY_VERSION=${trivy_version}
		WithEnvVariable("TRIVY_VERSION", TRIVYVERSION).
		WithExposedPort(8080).
		WithExposedPort(8443).
		WithEntrypoint([]string{"/home/scanner/bin/scanner-trivy"})
}

// internal function to build harbor-portal
func (m *Harbor) buildPortal(ctx context.Context, platform Platform) *dagger.Container {
	fmt.Println("🛠️  Building Harbor Portal...")

	swaggerYaml := dag.Container().From("alpine:latest").
		// for better caching
		WithMountedDirectory("/api", m.PortalSrc.Directory("./api")).
		WithWorkdir("/api").
		File("v2.0/swagger.yaml")

	LICENSE := dag.Container().From("alpine:latest").
		WithDirectory("/harbor", m.PortalSrc).
		WithWorkdir("/harbor").
		WithExec([]string{"ls"}).
		File("LICENSE")

	before := dag.Container().
		From("node:"+NODE_VERSION).
		WithMountedCache("/root/.bun/install/cache", dag.CacheVolume("bun")).
		WithMountedCache("/root/.npm", dag.CacheVolume("node")).
		WithMountedCache("/root/.angular", dag.CacheVolume("angular")).
		// for better caching
		WithDirectory("/harbor", m.PortalSrc).
		WithWorkdir("/harbor/src/portal").
		WithEnvVariable("NPM_CONFIG_REGISTRY", NPM_REGISTRY).
		WithEnvVariable("BUN_INSTALL_CACHE_DIR", "/root/.bun/install/cache").
		// $BUN_INSTALL_CACHE_DIR
		// WithExec([]string{"bun", "pm", "trust", "--all"}).
		WithFile("swagger.yaml", swaggerYaml).
		// WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "unzip"}).
		WithExec([]string{"npm", "install", "-g", "bun@" + BUN_VERSION}).
		WithExec([]string{"bun", "install"}).
		// WithExec([]string{"bun", "pm", "trust", "--all"}).
		// WithExec([]string{"bun", "install", "--no-verify"}).
		WithExec([]string{"ls", "-al"}).
		WithExec([]string{"bun", "run", "generate-build-timestamp"}).
		WithExec([]string{"bun", "run", "node", "--max_old_space_size=2048", "node_modules/@angular/cli/bin/ng", "build", "--configuration", "production"})

	builder := before.
		WithExec([]string{"bun", "install", "js-yaml@4.1.0", "--no-verify"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("bun -e \"const yaml = require('js-yaml'); const fs = require('fs'); const swagger = yaml.load(fs.readFileSync('swagger.yaml', 'utf8')); fs.writeFileSync('swagger.json', JSON.stringify(swagger));\" ")}).
		WithFile("/harbor/src/portal/dist/LICENSE", LICENSE)

	builderDir := builder.Directory("/harbor")

	// swagger UI only supports npm some edge case error
	swagger := dag.Container().
		From("node:"+NODE_VERSION).
		WithMountedCache("/root/.npm", dag.CacheVolume("node")).
		WithMountedCache("/root/.angular", dag.CacheVolume("angular")).
		WithMountedDirectory("/harbor", builderDir).
		WithWorkdir("/harbor/src/portal/app-swagger-ui").
		WithExec([]string{"npm", "install", "--unsafe-perm"}).
		WithExec([]string{"npm", "run", "build"}).
		WithWorkdir("/harbor/src/portal")

	deployer := dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(string(platform))}).From("nginx:alpine").
		WithDirectory("/etc/ssl/certs", m.getCaCerts(ctx)).
		WithFile("/usr/share/nginx/html/swagger.json", builder.File("/harbor/src/portal/swagger.json")).
		WithDirectory("/usr/share/nginx/html", builder.Directory("/harbor/src/portal/dist")).
		WithDirectory("/usr/share/nginx/html", swagger.Directory("/harbor/src/portal/app-swagger-ui/dist")).
		WithFile("/etc/nginx/nginx.conf", m.OnlyDagger.File("./.dagger/config/portal/nginx.conf")).
		WithWorkdir("/usr/share/nginx/html").
		WithExec([]string{"ls"}).
		WithWorkdir("/").
		WithExposedPort(8080).
		WithExposedPort(8443).
		WithEntrypoint([]string{"nginx", "-g", "daemon off;"})

	return deployer
}

// use to parse given platform as string
func parsePlatform(platform string) (string, string, error) {
	parts := strings.Split(platform, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid platform format: %s. Should be os/arch. E.g. darwin/amd64", platform)
	}
	return parts[0], parts[1], nil
}

// fetches git commit
func (m *Harbor) fetchGitCommit(ctx context.Context) string {
	dirOpts := dagger.ContainerWithDirectoryOpts{
		Include: []string{".git"},
	}

	// temp container with git installed
	temp := dag.Container().
		From("golang:"+GO_VERSION).
		WithDirectory("/src", m.FilteredSrc, dirOpts).
		WithWorkdir("/src")

	gitCommit, _ := temp.WithExec([]string{"git", "rev-parse", "--short=8", "HEAD"}).Stdout(ctx)

	return gitCommit
}

// generate APIs
func (m *Harbor) genAPIs(_ context.Context) *dagger.Directory {
	SWAGGER_SPEC := "api/v2.0/swagger.yaml"
	TARGET_DIR := "src/server/v2.0"
	APP_NAME := "harbor"

	temp := dag.Container().
		From("quay.io/goswagger/swagger:"+SWAGGER_VERSION).
		WithExec([]string{"swagger", "version"}).
		WithDirectory("/src", m.FilteredSrc).
		WithWorkdir("/src").
		// Clean up old generated code and create necessary directories
		// WithExec([]string{"rm", "-rf", TARGET_DIR + "/{models,restapi}"}).
		WithExec([]string{"mkdir", "-p", TARGET_DIR}).
		WithExec([]string{"ls", "-la"}).
		// Generate the server files using the Swagger tool
		WithExec([]string{"swagger", "generate", "server", "--template-dir=./tools/swagger/templates", "--exclude-main", "--additional-initialism=CVE", "--additional-initialism=GC", "--additional-initialism=OIDC", "-f", SWAGGER_SPEC, "-A", APP_NAME, "--target", TARGET_DIR}).
		// WithExec([]string{"ls", "-la"}).
		Directory("/src/src/server/v2.0")

	return temp
}

// get version from VERSION file
func (m *Harbor) GetVersion(ctx context.Context) string {
	dirOpts := dagger.ContainerWithDirectoryOpts{
		Include: []string{"VERSION"},
	}

	temp := dag.Container().
		From("golang:"+GO_VERSION+"-alpine").
		WithDirectory("/src", m.FilteredSrc, dirOpts).
		WithWorkdir("/src").
		WithExec([]string{"ls", "-la"})

	version, _ := temp.WithExec([]string{"cat", "VERSION"}).Stdout(ctx)
	return version
}

func (m *Harbor) getCaCerts(ctx context.Context) *dagger.Directory {
	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "ca-certificates"}).
		Directory("/etc/ssl/certs")
}
