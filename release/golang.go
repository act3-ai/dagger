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
}

// Go provides utilities for releasing golang projects.
func (r *Release) Go() *Golang {
	return &Golang{
		Release: r,
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
	// TODO: GOPRIVATE?
	res, err = dag.Go(dagger.GoOpts{Container: unitTestBase}).
		WithSource(g.Release.Source).
		Container().
		With(func(ctr *dagger.Container) *dagger.Container {
			if g.Release.Netrc != nil {
				ctr = ctr.WithMountedSecret("/root/.netrc", g.Release.Netrc)
			}
			return ctr
		}).
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
