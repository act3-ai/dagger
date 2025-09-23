package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"errors"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

type Golang struct {
	Release *Release

	GoPrivate string
}

// Go provides utilities for releasing golang projects.
func (r *Release) Go(
	// value of GOPRIVATE
	// +optional
	goPrivate string,
) *Golang {
	return &Golang{
		Release:   r,
		GoPrivate: goPrivate,
	}
}

// Check performs sanity checks prior to releasing.
//
// Specifically, it runs: shellcheck, yamllint, markdownlint-cli2, golangci-lint, govulncheck, and go unit tests.
func (g *Golang) Check(ctx context.Context,
	// base container for all linters.
	// +optional
	base *dagger.Container,
	// base container to run unit tests in.
	// +optional
	unitTestBase *dagger.Container,
	// skip any provided Generic lint tests
	// +optional
	skip []string,
) (string, error) {
	results := util.NewResultsBasicFmt(strings.Repeat("=", 15)) // must be concurrency safe

	p := pool.New().
		WithErrors().
		WithContext(ctx)

	p.Go(func(ctx context.Context) error {
		// shellcheck, yamllint, markdownlint
		_, err := g.Release.GenericLint(ctx, base, skip)
		if err != nil {
			return fmt.Errorf("running generic linters: %w", err)
		}
		results.Add("Generic Linters", "success")
		return nil
	})

	p.Go(func(ctx context.Context) error {
		// lint *.go
		res, err := dag.GolangciLint(dagger.GolangciLintOpts{Container: base}).
			Run(g.Release.Source, dagger.GolangciLintRunOpts{Timeout: "10m"}).
			Stdout(ctx)
		results.Add("Golangci-lint", res)
		if err != nil {
			return fmt.Errorf("running golangci-lint: %w", err)
		}
		return nil
	})

	p.Go(func(ctx context.Context) error {
		// govulncheck
		res, err := dag.Govulncheck(
			dagger.GovulncheckOpts{
				Container: base,
				Netrc:     g.Release.Netrc,
			}).
			ScanSource(ctx, g.Release.Source)
		results.Add("Govulncheck", res)
		if err != nil {
			return fmt.Errorf("running govulncheck: %w", err)
		}
		return nil
	})

	p.Go(func(ctx context.Context) error {
		// unit tests
		res, err := g.goContainer(unitTestBase).
			WithExec([]string{"go", "test", "./..."}).
			Stdout(ctx)
		results.Add("Go Unit Tests", res)
		if err != nil {
			return fmt.Errorf("running go unit tests: %w", err)
		}
		return nil
	})

	err := p.Wait()

	if errStatus := g.Release.gitStatus(ctx); errStatus != nil {
		err = errors.Join(err, fmt.Errorf("git repository is dirty after running linters and unit tests: %w", errStatus))
	}

	return results.String(), err
}

// Verify release version adheres to gorelease standards.
//
// See https://pkg.go.dev/golang.org/x/exp/cmd/gorelease.
func (g *Golang) Verify(ctx context.Context,
	// target module version.
	targetVersion string,
	// current module version. Default: contents of VERSION file, with 'v' prefix added.
	// +optional
	currentVersion string,
	// base container.
	// +optional
	base *dagger.Container,
) (string, error) {
	// const gorelease = "golang.org/x/exp/cmd/gorelease@latest"
	const gorelease = "github.com/nathan-joslin/exp/cmd/gorelease@d53ca235cbb4684a341c9f15f3e60fffe7c9f2c7"

	var err error
	if currentVersion == "" {
		currentVersion, err = g.Release.Source.File("VERSION").Contents(ctx)
		if err != nil {
			return "", fmt.Errorf("retreving version from VERSION file: %w", err)
		}
	}

	if !strings.HasPrefix(currentVersion, "v") {
		currentVersion = "v" + currentVersion
	}

	out, err := g.goContainer(nil).
		WithExec([]string{"go", "install", gorelease}).
		WithExec([]string{"/work/src/tool/gorelease",
			fmt.Sprintf("-base=%s", strings.TrimSpace(currentVersion)),
			fmt.Sprintf("-version=%s", strings.TrimSpace(targetVersion)),
		}).
		Stdout(ctx)

	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		// exit code != 0
		return out, fmt.Errorf("%s", e.Stderr)
	case err != nil:
		// some other dagger error, e.g. graphql
		return out, err
	default:
		// exit code = 0
		return out, nil
	}
}

// goContainer returns a go container with private auth setup, if applicable.
func (g *Golang) goContainer(base *dagger.Container) *dagger.Container {
	return dag.Go(dagger.GoOpts{Container: base}).
		WithSource(g.Release.Source).
		WithEnvVariable("GOBIN", "/work/src/tool").
		WithEnvVariable("GOPRIVATE", g.GoPrivate).
		Container().
		With(func(ctr *dagger.Container) *dagger.Container {
			if g.Release.Netrc != nil {
				ctr = ctr.WithMountedSecret("/root/.netrc", g.Release.Netrc)
			}
			return ctr
		}).
		With(func(c *dagger.Container) *dagger.Container {
			if g.Release.GitIgnore != nil {
				const gitIgnorePath = "/work/.gitignore"
				c = c.WithMountedFile(gitIgnorePath, g.Release.GitIgnore).
					WithExec([]string{"git", "config", "--global", "core.excludesfile", gitIgnorePath})
			}
			return c
		}).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod"))
}
