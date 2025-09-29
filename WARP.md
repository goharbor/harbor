# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

Project: Harbor (monorepo for building Harbor images, binaries, and installers)

Key repos/docs referenced in-repo:
- README.md: badges, install links, and architecture links
- Makefile: entrypoint for build, lint, packaging, and local compose orchestration
- src/ (Go, main services and libraries)
- make/ (Docker build context, photon-based images, compose, and scripts)
- tests/ (integration/Robot assets, compose for tests)

Common commands
- Build everything (compile Go, generate API, build Docker images, prepare config)
  make install
  # Equivalent to: make compile build prepare start

- Compile Go binaries with Dockerized toolchain (golang image) only
  make compile
  # Subtargets:
  make compile_core
  make compile_jobservice
  make compile_registryctl
  make compile_standalone_db_migrator

- Build Docker images (from make/photon/*)
  make build
  # Parameters you may override (examples):
  #   VERSIONTAG=dev|vX.Y.Z, BASEIMAGETAG=dev, BUILD_BASE=true|false, PULL_BASE_FROM_DOCKERHUB=true|false
  #   GOBUILDIMAGE=golang:1.24.6, NODEBUILDIMAGE=node:16.18.0, IMAGENAMESPACE=goharbor

- Start and stop local compose stack (uses make/docker-compose.yml)
  make start
  make down
  make restart

- Generate and validate API/mocks (required before some builds/lint)
  make lint_apis
  make gen_apis
  make gen_mocks
  make mocks_check

- Lint (golangci-lint; installs separately per comment in Makefile)
  make lint
  # Ensure $(go env GOPATH)/bin/golangci-lint exists or install as per Makefile comment

- Vulnerability scan for Go deps (requires govulncheck installed)
  make govulncheck

- Go checks bundle
  make go_check
  # runs: gen_apis, mocks_check, misspell, commentfmt, lint

- Package installers
  # Online installer
  make package_online PKGVERSIONTAG=<tag> [REGISTRYSERVER=host/ REGISTRYPROJECTNAME=project]
  # Offline installer (saves built images into tarball)
  make package_offline PKGVERSIONTAG=<tag>

- Push built images to registry
  make pushimage REGISTRYSERVER=<host/> REGISTRYUSER=<user> REGISTRYPASSWORD=****** DEVFLAG=false

- Clean
  make clean          # usage info for clean targets
  make cleanall       # remove binaries, images, compose files, configs, packages
  make cleanbinary    # remove compiled binaries
  make cleanimage     # remove Harbor images
  make cleanbaseimage # remove base images

- Run tests
  # Unit/integration Go tests live under src; use standard go test, typically within src
  (cd src && go test ./...)

  # Test compose and Robot-related flows are under tests/
  # Example helpers available:
  tests/integration.sh
  tests/startuptest.sh
  tests/swaggerchecker.sh
  # Single Go package test example:
  (cd src && go test ./pkg/trace -run TestSampler -v)

Notes on environment and prerequisites
- Docker and docker-compose are required for build/start/package flows (README lists minimum versions).
- Build largely orchestrated via photon-based Dockerfiles in make/photon; builds run in containers to avoid host toolchain drift.
- VERSION holds the release version (current: contents of VERSION file; e.g., v2.15.0). VERSIONTAG defaults to dev for images.
- Some targets use network access to fetch swagger/mocks tool images and dependencies.
- Setting GEN_TLS=1 when running make prepare can generate internal TLS certs via prepare image.

High-level architecture
Harbor in this repo builds a multi-service system packaged as containers and orchestrated with docker-compose or Helm (chart is external). The main runtime services and their in-repo code live under src/:
- core (src/core):
  - Web/API service implementing Harbor business logic and REST APIs (v2.0 swagger at api/v2.0/swagger.yaml with generated server under src/server/v2.0)
  - Controllers, middlewares, auth, session mgmt, and Beego-based components
- jobservice (src/jobservice):
  - Asynchronous job orchestration built on Redis (gocraft/work fork)
  - Handles replication, GC, scan, notifications, exports, system artifact cleanup, etc.
  - Components: API server, controller, scheduler, logger, reaper, Redis-backed queue; supports hooks and status web UI endpoints
- registryctl (src/registryctl):
  - Sidecar/controller for upstream distribution registry (patched) for access control and cleanup; interacts with core and registry
- portal (src/portal):
  - UI frontend; built via Node image, artifacts bundled into portal container
- common libs (src/lib, src/pkg, src/common):
  - Cross-cutting libraries: config, http, errors, cache/redis wrappers, metrics, tracing (OpenTelemetry), icon, etc.
  - Domain packages under src/pkg (retention, scanner, p2p preheat, scheduler, task, queuestatus, etc.)
- migration (src/migration):
  - DB and data model migrations
- server (src/server/v2.0):
  - Swagger-generated API server glue for core
- cmd (src/cmd):
  - Auxiliary binaries, e.g., standalone DB migrator, exporter

Build and image pipeline
- make compile_* builds Go binaries inside a golang container, embedding version info via -ldflags with src/pkg/version (GitCommit and ReleaseVersion from VERSION file and current commit).
- make build delegates to make/photon/Makefile to build base images and service images: core, jobservice, registry, registryctl, nginx, portal, db, redis, exporter, and prepare.
- make prepare renders config (harbor.yml) and compose files via the prepare image and make/ scripts; can generate internal TLS if GEN_TLS is set.
- make start/up uses docker-compose.yml under make/ to bring up the composed services.

API and codegen
- Swagger spec: api/v2.0/swagger.yaml
- Lint spec: make lint_apis (Spectral via tools/spectral Docker image)
- Generate server stubs: make gen_apis (OpenAPI generator via tools/swagger Docker image; outputs to src/server/v2.0)
- Mocks: make gen_mocks (mockery via tools/mockery Docker image); make mocks_check ensures mocks are up to date

Testing strategy (in-repo signals)
- Go unit/integration: standard go test under src/ (golangci config at src/.golangci.yaml)
- Robot/Selenium style UI tests and helpers under tests/, with supporting docker-compose.test.yml
- Helper scripts in tests/: integration.sh, startuptest.sh, swaggerchecker.sh, etc.

Service secrets and configuration
- make prepare and prepare image manage assembling harbor.yml, secrets, and compose files under make/; many config knobs are parameterized in Helm chart (external) and values test data under src/pkg/chart/testdata/ for reference.

Working tips specific to this repo
- Always run gen_apis and gen_mocks (or simply make go_check) after editing swagger or interfaces in src/pkg to keep generated code/mocks in sync; CI checks for drift.
- Use containerized build targets (default) to avoid local Go/Node version mismatches. Override GOBUILDIMAGE/NODEBUILDIMAGE only if necessary.
- For local dev on Go packages: run go test ./... from src/ to avoid pulling the entire monorepo module graph at top-level (module root is src/ per src/go.mod with replace to ../).
