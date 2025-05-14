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

// TODO: consider adding the release string formmatter to release struct itself
// TODO: helm chart version bumping, make it flexible to zero or more helm charts
// TODO: add support for modifications to releases.md for images and helm chart table

// Check performs sanity checks prior to releasing, i.e. linters and unit tests.
func (r *Release) Check(ctx context.Context) (string, error) {
	if err := r.gitStatus(ctx); err != nil {
		return "", fmt.Errorf("git repository is dirty, aborting check: %w", err)
	}

	results := util.NewResultsBasicFmt(strings.Repeat("=", 15))

	if err := r.genericLint(ctx, results); err != nil {
		return results.String(), fmt.Errorf("running generic linters: %w", err)
	}

	if err := r.checkByProjectType(ctx, results); err != nil {
		return results.String(), fmt.Errorf("preparing based on project type %s: %w", r.ProjectType, err)
	}

	return results.String(), nil
}

// checkByProjectType performs language specific checks.
func (r *Release) checkByProjectType(ctx context.Context, results util.ResultsFormatter) error {
	switch r.ProjectType {
	case util.Golang:
		return r.checkGolang(ctx, results)
	case util.Python:
		return r.checkPython(ctx, results)
	default:
		// sanity, should be impossible
		return fmt.Errorf("unsupported project type %s", r.ProjectType)
	}
}

// checkGolang runs go specific checks.
func (r *Release) checkGolang(ctx context.Context, results util.ResultsFormatter) error {
	var errs []error

	// lint
	res, err := dag.GolangciLint().
		Run(r.Source, dagger.GolangciLintRunOpts{Timeout: "10m"}).
		Stdout(ctx)
	results.Add("Golangci-lint", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running golangci-lint: %w", err))
	}

	// govulncheck
	res, err = dag.Govulncheck().
		With(func(v *dagger.Govulncheck) *dagger.Govulncheck {
			if r.Netrc != nil {
				v = v.WithNetrc(r.Netrc)
			}
			return v
		}).
		ScanSource(r.Source).
		Stdout(ctx)
	results.Add("Govulncheck", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running govulncheck: %w", err))
	}

	// unit tests
	if !r.DisableUnitTests {
		res, err = dag.Go().
			WithSource(r.Source).
			Container().
			With(func(ctr *dagger.Container) *dagger.Container {
				if r.Netrc != nil {
					ctr = ctr.WithMountedSecret("/root/.netrc", r.Netrc)
				}
				return ctr
			}).
			WithExec([]string{"go", "test", "./..."}).
			Stdout(ctx)
		results.Add("Go Unit Tests", res)
		if err != nil {
			errs = append(errs, fmt.Errorf("running go unit tests: %w", err))
		}
	}

	// TODO: go generate?

	return errors.Join(errs...)
}

// checkPython runs python specific preparations.
func (r *Release) checkPython(ctx context.Context, results util.ResultsFormatter) error {
	// python linters
	dag.Python(
		dagger.PythonOpts{
			Src:   r.Source,
			Netrc: r.Netrc,
		},
	).
		Test()

	// ...

	return fmt.Errorf("not implemented")
}

// genericLint runs geneic linters, e.g. markdown, yaml, etc.
func (r *Release) genericLint(ctx context.Context, results util.ResultsFormatter) error {
	var errs []error

	res, err := r.shellcheck(ctx, 4) // TODO: plumb concurrency?
	results.Add("Shellcheck", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running shellcheck: %w", err))
	}

	res, err = dag.Yamllint().
		Run(r.Source).
		Stdout(ctx)
	results.Add("Yamllint", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running yamllint: %w", err))
	}

	res, err = dag.Markdownlint().
		Run(r.Source, []string{"."}).
		Stdout(ctx)
	results.Add("Markdownlint", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running markdownlint: %w", err))
	}

	return errors.Join(errs...)
}

// shellcheck auto-detects and runs on all *.sh and *.bash files in the source directory.
//
// Users who want custom functionality should use github.com/dagger/dagger/modules/shellcheck directly.
func (r *Release) shellcheck(ctx context.Context, concurrency int) (string, error) {

	// TODO: Consider adding an option for specifying script files that don't have the extension, such as WithShellScripts.
	shEntries, err := r.Source.Glob(ctx, "**/*.sh")
	if err != nil {
		return "", fmt.Errorf("globbing shell scripts with *.sh extension: %w", err)
	}

	bashEntries, err := r.Source.Glob(ctx, "**/*.bash")
	if err != nil {
		return "", fmt.Errorf("globbing shell scripts with *.bash extension: %w", err)
	}

	p := pool.NewWithResults[string]().
		WithMaxGoroutines(concurrency).
		WithErrors().
		WithContext(ctx)

	entries := append(shEntries, bashEntries...)
	for _, entry := range entries {
		p.Go(func(ctx context.Context) (string, error) {
			r, err := dag.Shellcheck().
				Check(r.Source.File(entry)).
				Report(ctx)
			if r == "" {
				r = "No reported issues."
			}
			r = fmt.Sprintf("Results for file %s:\n%s", entry, r)
			return r, err
		})
	}

	res, err := p.Wait()
	return strings.Join(res, "\n\n"), err
}

// gitStatus returns an error if a git repository contains uncommitted changes.
func (r *Release) gitStatus(ctx context.Context) error {
	ctr := dag.Wolfi().
		Container(
			dagger.WolfiContainerOpts{
				Packages: []string{"git"},
			},
		).
		WithMountedDirectory("/work/src", r.Source).
		WithWorkdir("/work/src")

	var errs []error

	// check for unstaged changes
	_, err := ctr.WithExec([]string{"git", "diff", "--stat", "--exit-code"}).Stdout(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("checking for unstaged git changes: %w", err))
	}

	// check for staged, but not committed changes
	_, err = ctr.WithExec([]string{"git", "diff", "--cached", "--stat", "--exit-code"}).Stdout(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("checking for staged git changes: %w", err))
	}
	return errors.Join(errs...)
}
