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

	pl := pool.New().
		WithErrors().
		WithContext(ctx)

	pl.Go(func(ctx context.Context) error {
		// shellcheck, yamllint, markdownlint
		_, err := p.Release.GenericLint(ctx, Base)
		if err != nil {
			return fmt.Errorf("running generic linters: %w", err)
		}
		results.Add("Generic Linters", "success")
		return nil
	})

	pl.Go(func(ctx context.Context) error {
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
			return fmt.Errorf("Python Linting Error: %w", err)
		} else {
			results.Add("Python Lint", "Success")
			return nil
		}
	})

	err := pl.Wait()

	if errStatus := p.Release.gitStatus(ctx); errStatus != nil {
		err = errors.Join(err, fmt.Errorf("git repository is dirty after running linters and unit tests: %w", errStatus))
	}

	return results.String(), err
}
