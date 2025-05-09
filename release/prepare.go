package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

// TODO: consider adding the release string formmatter to release struct itself
// TODO: helm chart version bumping, make it flexible to zero or more helm charts
// TODO: add support for modifications to releases.md for images and helm chart table

// Prepare performs sanity checks prior to releasing.
func (r *Release) Prepare(ctx context.Context) (string, error) {
	results := util.NewResultsBasicFmt(strings.Repeat("=", 15))

	if err := r.genericLint(ctx, results); err != nil {
		return results.String(), fmt.Errorf("running generic linters: %w", err)
	}

	if err := r.prepByProjectType(ctx, results); err != nil {
		return results.String(), fmt.Errorf("preparing based on project type %s: %w", r.ProjectType, err)
	}

	return "", fmt.Errorf("not implemented")
}

// prepareByProjectType performs language specific preparations.
func (r *Release) prepByProjectType(ctx context.Context, results util.ResultsFormatter) error {
	switch r.ProjectType {
	case util.Golang:
		return r.prepGolang(ctx, results)
	case util.Python:
		return r.prepPython(ctx, results)
	default:
		// sanity, should be impossible
		return fmt.Errorf("unsupported project type %s", r.ProjectType)
	}
}

// prepGolang runs go specific preparations.
func (r *Release) prepGolang(ctx context.Context, results util.ResultsFormatter) error {
	var errs []error

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

	// TODO: go generate

	return errors.Join(errs...)
}

// prepPython runs python specific preparations.
func (r *Release) prepPython(ctx context.Context, results util.ResultsFormatter) error {
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
	srcFiltered := r.Source.Filter(
		dagger.DirectoryFilterOpts{
			Include: []string{"*.sh", "*.bash"}, // only supports bash/sh
		},
	)

	entries, err := srcFiltered.Entries(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving entries from filtered source directory: %w", err)
	}

	p := pool.NewWithResults[string]().
		WithMaxGoroutines(concurrency).
		WithErrors().
		WithContext(ctx)
	for _, entry := range entries {
		fi, err := os.Stat(entry)
		if err != nil {
			return "", fmt.Errorf("retrieving file info for %s: %w", entry, err)
		}
		if fi.IsDir() {
			continue
		}

		p.Go(func(ctx context.Context) (string, error) {
			r, err := dag.Shellcheck().
				Check(srcFiltered.File(entry)).
				Report(ctx)
			r = fmt.Sprintf("Results for file %s:\n%s", entry, r)
			return r, err
		})
	}

	res, err := p.Wait()
	return strings.Join(res, "\n\n"), err
}
