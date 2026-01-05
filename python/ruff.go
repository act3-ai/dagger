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
	// +private
	Changes *dagger.Changeset
}

// contains commands for running ruff on a Python project.
func (p *Python) Ruff() *Ruff {
	return &Ruff{Python: p}
}

// Runs ruff check and returns a container that will fail on any errors.
func (r *Ruff) Lint(
	// +optional
	// +default="full"
	outputFormat string,
) *dagger.Container {
	// Run ruff check with the provided output format
	return r.Python.Container().WithExec(
		[]string{
			"uv",
			"run",
			"--with=ruff",
			"ruff",
			"check", ".",
			"--output-format", outputFormat})

}

// Runs ruff check and returns a results in a json file.
func (r *Ruff) Report() *dagger.File {
	// Run ruff check with the provided output format
	return r.Python.Container().WithExec(
		[]string{
			"uv",
			"run",
			"--with=ruff",
			"ruff",
			"check", ".",
			"--output-format",
			"json",
			"--output-file",
			"ruff-results.json"},
		dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}).
		File("ruff-results.json")

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
