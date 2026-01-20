package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/goact-stack/internal/dagger"
)

type GoactStack struct{}

const embedPlaceholder = `<!DOCTYPE html><html><head><title>Build Required</title></head><body><h1>Run bun run build in frontend/ first</h1></body></html>`

// withEmbedPlaceholder ensures the static embed directory has a file so Go builds succeed.
// This avoids requiring a frontend build just to lint or test Go code.
func (m *GoactStack) withEmbedPlaceholder(source *dagger.Directory) *dagger.Directory {
	return source.WithNewFile("internal/static/dist/index.html", embedPlaceholder)
}

func (m *GoactStack) gitVersion(ctx context.Context, git *dagger.Directory) (string, error) {
	if git == nil {
		return "dev", nil
	}
	out, err := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/src/.git", git).
		WithWorkdir("/src").
		WithExec([]string{"git", "describe", "--tags", "--always"}).
		Stdout(ctx)
	if err != nil {
		return "dev", nil
	}
	return strings.TrimSpace(out), nil
}

func (m *GoactStack) Version(
	ctx context.Context,
	// +optional
	// +defaultPath="/.git"
	git *dagger.Directory,
) (string, error) {
	return m.gitVersion(ctx, git)
}

func (m *GoactStack) Build(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +defaultPath="/.git"
	git *dagger.Directory,
	// +optional
	version string,
) (*dagger.Container, error) {
	if version == "" {
		v, err := m.gitVersion(ctx, git)
		if err != nil {
			return nil, fmt.Errorf("version detection failed: %w", err)
		}
		version = v
	}

	if _, err := m.lintSource(ctx, source); err != nil {
		return nil, fmt.Errorf("lint failed: %w", err)
	}

	if _, err := m.testSource(ctx, source); err != nil {
		return nil, fmt.Errorf("test failed: %w", err)
	}

	// Build frontend assets
	assetsDir := m.BuildFrontend(source)
	buildSource := source.WithDirectory("internal/static/dist", assetsDir)

	return m.BuildBinary(buildSource, version), nil
}

func (m *GoactStack) Lint(ctx context.Context, source *dagger.Directory) (string, error) {
	return m.lintSource(ctx, source)
}

func (m *GoactStack) lintSource(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.Container().
		From("golangci/golangci-lint:v2.8.0-alpine").
		WithEnvVariable("GOCACHE", "/go-build-cache").
		WithEnvVariable("GOMODCACHE", "/go-mod-cache").
		WithEnvVariable("GOLANGCI_LINT_CACHE", "/golangci-lint-cache").
		WithMountedCache("/go-build-cache", dag.CacheVolume("go-build-cache")).
		WithMountedCache("/go-mod-cache", dag.CacheVolume("go-mod-cache")).
		WithMountedCache("/golangci-lint-cache", dag.CacheVolume("golangci-lint-cache")).
		WithDirectory("/app", m.withEmbedPlaceholder(source)).
		WithWorkdir("/app").
		WithExec([]string{"golangci-lint", "run", "--timeout", "5m"}).
		Stdout(ctx)
}

func (m *GoactStack) Typecheck(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.Container().
		From("ghcr.io/jwhumphries/frontend:latest").
		WithDirectory("/app", source).
		WithWorkdir("/app/frontend").
		WithExec([]string{"bun", "install"}).
		WithExec([]string{"bun", "run", "typecheck"}).
		Stdout(ctx)
}

func (m *GoactStack) LintFrontend(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.Container().
		From("ghcr.io/jwhumphries/frontend:latest").
		WithDirectory("/app", source).
		WithWorkdir("/app/frontend").
		WithExec([]string{"bun", "install"}).
		WithExec([]string{"bun", "run", "lint"}).
		Stdout(ctx)
}

func (m *GoactStack) Test(ctx context.Context, source *dagger.Directory) (string, error) {
	return m.testSource(ctx, source)
}

func (m *GoactStack) testSource(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.Container().
		From("golang:1.25-alpine").
		WithEnvVariable("GOCACHE", "/go-build-cache").
		WithEnvVariable("GOMODCACHE", "/go-mod-cache").
		WithMountedCache("/go-build-cache", dag.CacheVolume("go-build-cache")).
		WithMountedCache("/go-mod-cache", dag.CacheVolume("go-mod-cache")).
		WithDirectory("/app", m.withEmbedPlaceholder(source)).
		WithWorkdir("/app").
		WithExec([]string{"go", "test", "-v", "./..."}).
		Stdout(ctx)
}

// BuildFrontend compiles the React/TypeScript frontend with Vite
func (m *GoactStack) BuildFrontend(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("ghcr.io/jwhumphries/frontend:latest").
		WithDirectory("/app", source).
		WithWorkdir("/app/frontend").
		WithExec([]string{"bun", "install"}).
		WithExec([]string{"bun", "run", "build"}).
		Directory("/app/internal/static/dist")
}

func (m *GoactStack) BuildBinary(source *dagger.Directory, version string) *dagger.Container {
	return dag.Container().
		From("golang:1.25-alpine").
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithEnvVariable("GOCACHE", "/go-build-cache").
		WithEnvVariable("GOMODCACHE", "/go-mod-cache").
		WithMountedCache("/go-build-cache", dag.CacheVolume("go-build-cache")).
		WithMountedCache("/go-mod-cache", dag.CacheVolume("go-mod-cache")).
		WithExec([]string{
			"go", "build",
			"-ldflags", "-X goact-stack/version.Tag=" + version,
			"-o", "/goact-stack",
			"./cmd/goact-stack/",
		})
}

func (m *GoactStack) Release(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +defaultPath="/.git"
	git *dagger.Directory,
	// +optional
	version string,
) (*dagger.Container, error) {
	binaryContainer, err := m.Build(ctx, source, git, version)
	if err != nil {
		return nil, err
	}
	binary := binaryContainer.File("/goact-stack")

	return dag.Container().
		From("alpine:3.21").
		WithExec([]string{"apk", "add", "--no-cache", "tzdata", "ca-certificates"}).
		WithFile("/usr/local/bin/goact-stack", binary).
		WithExec([]string{"sh", "-c", "echo 'nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin' >> /etc/passwd"}).
		WithEnvVariable("TZ", "America/New_York").
		WithEnvVariable("GOACT_PORT", ":8080").
		WithExposedPort(8080).
		WithUser("10001").
		WithEntrypoint([]string{"/usr/local/bin/goact-stack"}), nil
}

func (m *GoactStack) Fmt(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("golang:1.25-alpine").
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec([]string{"go", "fmt", "./..."}).
		Directory("/app")
}

func (m *GoactStack) FmtFrontend(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("ghcr.io/jwhumphries/frontend:latest").
		WithDirectory("/app", source).
		WithWorkdir("/app/frontend").
		WithExec([]string{"bun", "install"}).
		WithExec([]string{"bun", "run", "lint", "--fix"}).
		Directory("/app")
}
