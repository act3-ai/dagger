package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
)

type Ruff struct {
	// +private
	Python *Python
}
type RuffCheckResults struct {
	// returns results of ruff-check as a file
	Results *dagger.File
	// returns exit code of pyright
	ExitCode int
}

// Runs ruff check on a given source directory. Returns a results file and an exit-code.
func (r *Ruff) Check(ctx context.Context,
	// +optional
	// +default="full"
	outputFormat string,
) (*RuffCheckResults, error) {
	// Run ruff check with the provided output format
	ctr, err := r.Python.Container().WithExec(
		[]string{
			"uv",
			"run",
			"--with=ruff",
			"ruff",
			"check", ".",
			"--output-format", outputFormat}, dagger.ContainerWithExecOpts{
			RedirectStdout: "/ruffcheck-results.txt",
			Expect:         dagger.ReturnTypeAny}).
		Sync(ctx)

	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running ruff-check: %w", err)
	}

	results := ctr.File("/ruffcheck-results.txt")

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	return &RuffCheckResults{
		Results:  results,
		ExitCode: exitCode,
	}, nil

}

// Runs ruff format against a given source directory.
// Returns a Changeset that can be used to apply any changes found
// to the host.
func (r *Ruff) Format(ctx context.Context,
	// file pattern to exclude from ruff format
	// +optional
	exclude []string) (*dagger.Changeset, error) {
	args := []string{
		"uv",
		"run",
		"--with=ruff",
		"ruff",
		"format",
		".",
	}

	// exclude any given file patterns
	if len(exclude) != 0 {
		for _, exclude := range exclude {
			args = append(args, "--exclude", exclude)
		}
	}

	ctr, err := r.Python.Container().
		WithExec(args).
		Sync(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running ruff-format: %w", err)
	}

	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv", ".ruff_cache"}})

	return afterChanges.Changes(r.Python.Source), nil
}
