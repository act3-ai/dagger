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
func (p *Py) Check(ctx context.Context,
	// base container for tests can be overwritten
	// +optional
	Base *dagger.Container,
	// extra arguments for uv sync command
	// +optional
	SyncArgs []string,
	// unit test directory
	// +optional
	// +default="test"
	UnitTestDir string,
	// skip any provided lint tests
	// +optional
	skip []string,
) (string, error) {
	results := util.NewResultsBasicFmt(strings.Repeat("=", 15))
	var errs []error
	if err := p.Release.genericLint(ctx, results, Base); err != nil {
		errs = append(errs, fmt.Errorf("running generic linters: %w", err))
	}

	// python linters
	_, err := dag.Python(p.Release.Source,
		dagger.PythonOpts{
			Base:     Base,
			Netrc:    p.Release.Netrc,
			SyncArgs: SyncArgs,
		},
	).
		Test(dagger.PythonTestOpts{
			UnitTestDir: UnitTestDir,
			Skip:        skip,
		}).Sync(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("Python Linting Error: %w", err))
	} else {
		results.Add("Python Lint", "Success")
	}

	return results.String(), errors.Join(errs...)
}
