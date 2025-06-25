package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"errors"
	"fmt"
	"strings"
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
) (string, error) {
	var errs []error
	results := util.NewResultsBasicFmt(strings.Repeat("=", 15))

	if err := g.Release.genericLint(ctx, results, base); err != nil {
		errs = append(errs, fmt.Errorf("running generic linters: %w", err))
	}

	// lint *.go
	res, err := dag.GolangciLint(dagger.GolangciLintOpts{Container: base}).
		Run(g.Release.Source, dagger.GolangciLintRunOpts{Timeout: "10m"}).
		Stdout(ctx)
	results.Add("Golangci-lint", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running golangci-lint: %w", err))
	}

	// govulncheck
	res, err = dag.Govulncheck(
		dagger.GovulncheckOpts{
			Container: base,
			Netrc:     g.Release.Netrc,
		}).
		ScanSource(ctx, g.Release.Source)
	results.Add("Govulncheck", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running govulncheck: %w", err))
	}

	// unit tests
	res, err = g.goContainer(unitTestBase).
		WithExec([]string{"go", "test", "./..."}).
		Stdout(ctx)
	results.Add("Go Unit Tests", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running go unit tests: %w", err))
	}

	if err := g.Release.gitStatus(ctx); err != nil {
		errs = append(errs, fmt.Errorf("git repository is dirty, aborting check: %w", err))
	}

	return results.String(), errors.Join(errs...)
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
) error {
	const gorelease = "golang.org/x/exp/cmd/gorelease@latest"

	var err error
	if currentVersion == "" {
		currentVersion, err = g.Release.Source.File("VERSION").Contents(ctx)
		if err != nil {
			return fmt.Errorf("retreving version from VERSION file: %w", err)
		}
	}

	if !strings.HasPrefix(currentVersion, "v") {
		currentVersion = "v" + currentVersion
	}

	_, err = g.goContainer(nil).
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
		return fmt.Errorf("%s", e.Stderr)
	case err != nil:
		// some other dagger error, e.g. graphql
		return err
	default:
		// exit code = 0
		return nil
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
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod"))
}
