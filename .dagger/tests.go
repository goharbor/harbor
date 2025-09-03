package main

import (
	"context"
	"dagger/harbor/internal/dagger"
	"fmt"
)

// Executes Linter and writes results to a file golangci-lint.report for github-actions
func (m *Harbor) LintReport(ctx context.Context) *dagger.File {
	report := "golangci-lint.report"
	return m.lint(ctx).WithExec([]string{
		"golangci-lint", "-v", "run", "--timeout=10m",
		"--out-format", "github-actions:" + report,
		"--issues-exit-code", "0",
	}).File(report)
}

// Lint Run the linter golangci-lint
func (m *Harbor) Lint(ctx context.Context) (string, error) {
	return m.lint(ctx).WithExec([]string{"golangci-lint", "-v", "run", "--timeout=10m"}).Stderr(ctx)
}

func (m *Harbor) lint(ctx context.Context) *dagger.Container {
	fmt.Println("👀 Running linter.")
	m.lintAPIs(ctx).Sync(ctx)
	m.mocksCheck(ctx).Sync(ctx)
	m.Source = m.genAPIs(ctx)
	linter := dag.Container().
		From("golangci/golangci-lint:"+GOLANGCILINT_VERSION+"-alpine").
		WithMountedCache("/lint-cache", dag.CacheVolume("/lint-cache")).
		WithEnvVariable("GOLANGCI_LINT_CACHE", "/lint-cache").
		WithMountedDirectory("/harbor", m.Source).
		WithWorkdir("/harbor/src")
		// WithExec([]string{"golangci-lint", "cache", "clean"})

	return linter
}

func (m *Harbor) goVulnCheck(ctx context.Context) *dagger.Container {
	m.Source = m.genAPIs(ctx)
	return dag.Container().
		From("golang:alpine").
		WithMountedDirectory("/harbor", m.Source).
		WithWorkdir("/harbor/src").
		WithExec([]string{"go", "install", "golang.org/x/vuln/cmd/govulncheck@latest"}).
		WithEntrypoint([]string{"/go/bin/govulncheck"})
}

// Check vulnerabilities in go-code
func (m *Harbor) GoVulnCheck(ctx context.Context) (string, error) {
	fmt.Println("👀 Running Go vulnerabilities check")
	return m.goVulnCheck(ctx).WithExec([]string{"govulncheck", "-show", "verbose", "./..."}).Stdout(ctx)
}

// Generate Vulnerability Report in sarif format for github-actions
func (m *Harbor) GoVulnCheckReport(ctx context.Context) (string, error) {
	fmt.Println("👀 Generating Vulnerability Report")
	return m.goVulnCheck(ctx).WithExec([]string{"govulncheck", "-format", "sarif", "./..."}).Stdout(ctx)
}

func (m *Harbor) lintAPIs(_ context.Context) *dagger.Directory {
	temp := dag.Container().
		From("stoplight/spectral:"+ SPECTRAL_VERSION).
		WithMountedDirectory("/src", m.Source).
		WithWorkdir("/src").
		WithExec([]string{"spectral", "--version"}).
		WithExec([]string{"spectral", "lint", "./api/v2.0/swagger.yaml"}).
		Directory("/src")

	return temp
}

// Check for outdated mocks
func (m *Harbor) mocksCheck(_ context.Context) *dagger.Directory {
	// script to check if mocks are outdated
	script := `
    res=$(git status -s src/ | awk '{ printf("%s\n", $2) }' | egrep .*.go)
    if [ -n "$res" ]; then
      echo "Mocks of the interface are out of date..."
      echo "$res"
      exit 1
    fi
	`

	return dag.Container().From("golang:latest").
		WithMountedDirectory("/src", m.Source).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", script}).
		Directory("/src")
}
