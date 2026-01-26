package main

import (
	"dagger/python/internal/dagger"
)

type Ruff struct {
	// +private
	Python *Python
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

// Runs ruff format against a given source directory. Returns a Changeset
// that can be used to apply any changes found
// to the host.
func (r *Ruff) Fix(
	// file pattern to exclude from ruff format
	// +optional
	exclude []string) *dagger.Changeset {
	args := []string{
		"uv",
		"run",
		"--with=ruff",
		"ruff",
		"format",
		".",
	}

	// exclude any given file patterns
	for _, exclude := range exclude {
		args = append(args, "--exclude", exclude)
	}

	ctr := r.Python.Container().
		WithExec(args)

	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv", ".ruff_cache"}})

	return afterChanges.Changes(r.Python.Base.Directory("/app"))
}
