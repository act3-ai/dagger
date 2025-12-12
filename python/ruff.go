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
type RuffLintResults struct {
	// returns results of ruff lint as a file
	Results *dagger.File
	// returns exit code of ruff lint
	// +private
	ExitCode int
}

type RuffFormatResults struct {
	Changes *dagger.Changeset
}

// run ruff commands on a given source directory.
func (p *Python) Ruff() *Ruff {
	return &Ruff{Python: p}
}

// Runs ruff check on a given source directory. Returns a results file and an exit-code.
func (r *Ruff) Lint(ctx context.Context,
	// +optional
	// +default="full"
	outputFormat string,
) (*RuffLintResults, error) {
	// Run ruff check with the provided output format
	ctr, err := r.Python.Container().WithExec(
		[]string{
			"uv",
			"run",
			"--with=ruff",
			"ruff",
			"check", ".",
			"--output-format", outputFormat}, dagger.ContainerWithExecOpts{
			RedirectStdout: "/rufflint-results.txt",
			Expect:         dagger.ReturnTypeAny}).
		Sync(ctx)

	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running ruff-check: %w", err)
	}

	results := ctr.File("/rufflint-results.txt")

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	return &RuffLintResults{
		Results:  results,
		ExitCode: exitCode,
	}, nil

}

// Check for any errors running ruff lint
func (rl *RuffLintResults) Check(ctx context.Context) error {
	if rl.ExitCode == 0 {
		return nil
	}
	results, err := rl.Results.Contents(ctx)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", results)
}

// Runs ruff format against a given source directory.
// Returns a Changeset that can be used to apply any changes found
// to the host.
func (r *Ruff) Format(ctx context.Context,
	// file pattern to exclude from ruff format
	// +optional
	exclude []string) (*RuffFormatResults, error) {
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

	return &RuffFormatResults{
		Changes: afterChanges.Changes(r.Python.Source),
	}, nil
}

// returns the results of ruff format as a changeset that can be applied to the host.
func (r *RuffFormatResults) Fix() (*dagger.Changeset, error) {
	return r.Changes, nil
}

// Returns an error if ruff format made any changes
func (r *RuffFormatResults) Check(ctx context.Context) error {
	empty, err := r.Changes.IsEmpty(ctx)
	if err != nil {
		return err
	}

	if empty {
		return nil
	}

	diff, err := r.Changes.AsPatch().Contents(ctx)
	if err != nil {
		return err
	}

	return fmt.Errorf("ruff format changes found:\n%s", diff)
}
