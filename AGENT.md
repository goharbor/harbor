# Harbor Agent Configuration

## Build/Test Commands
- **Build**: `make build` (full build), `make compile` (Go binaries only)
- **Tests**: `make go_check` (Go lint + tests), `go test ./src/...` (Go unit tests), `go test -run TestName ./src/pkg/package` (single test)
- **Lint**: `make lint` (Go with golangci-lint), `cd src/portal && npm run lint` (Angular)
- **Angular**: `cd src/portal && npm run test` (tests), `npm run test:headless` (CI), `npm run start` (dev server)
- **Coverage**: `./tests/coverage4gotest.sh` (Go), Angular tests include coverage by default

## Architecture
Harbor is a cloud-native container registry with microservices architecture:
- **Core**: Main API server (Go/Beego), handles authentication, projects, artifacts
- **Portal**: Angular 16 frontend with VMware Clarity Design System
- **JobService**: Background job processing (Go), replication, scanning, garbage collection  
- **Registry**: Docker registry v2 API implementation
- **Database**: PostgreSQL 15 with Redis cache
- **Components**: NGINX (proxy), Trivy (vulnerability scanning), Notary (image signing)

## Code Style & Conventions
- **Go**: Use `golangci-lint`, import order: stdlib, 3rd party, `github.com/goharbor/harbor`
- **Error handling**: Return wrapped errors, use `pkg/errors` package
- **Naming**: camelCase for Go, kebab-case for Angular components
- **Testing**: Place tests in `*_test.go` files, use testify/assert for Go, Jasmine for Angular
- **Format**: `gofmt` and `goimports` for Go, Prettier + ESLint for Angular
- **Headers**: All files require copyright headers (see `copyright.tmpl`)
