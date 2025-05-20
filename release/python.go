package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"errors"
	"fmt"
	"strings"
)

type Py struct {
	// +private
	Release *Release
}

func (r *Release) Python() *Py {
	return &Py{
		Release: r,
	}
}

// Check performs sanity checks prior to releasing, i.e. linters and unit tests.
func (p *Py) Check(ctx context.Context) (string, error) {
	results := util.NewResultsBasicFmt(strings.Repeat("=", 15))
	var errs []error
	if err := p.Release.genericLint(ctx, results); err != nil {
		errs = append(errs, fmt.Errorf("running generic linters: %w", err))
	}

	// python linters
	_, err := dag.Python(
		dagger.PythonOpts{
			Src:   p.Release.Source,
			Netrc: p.Release.Netrc,
		},
	).
		Test().Sync(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("Python Linting Error: %w", err))
	} else {
		results.Add("Python Lint", "Success")
	}

	if err := p.Release.gitStatus(ctx); err != nil {
		errs = append(errs, fmt.Errorf("git repository is dirty, aborting check: %w", err))
	}

	return results.String(), errors.Join(errs...)
}
