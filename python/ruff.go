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
	return r.Python.Container().
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=ruff",
				"ruff",
				"check", ".",
				"--output-format", outputFormat})

}

// Runs ruff check and attempts to fix any lint errors. Returns a changeset
// that can be used to apply any changes found
// to the host. Will return an error if any errors found are not considered fixable by ruff.
func (r *Ruff) LintFix(
	// +optional
	// +default="full"
	outputFormat string,
) *dagger.Changeset {
	// Run ruff check with the provided output format
	ctr := r.Python.Container().
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=ruff",
				"ruff",
				"check", ".",
				"--output-format", outputFormat,
				"--fix"})

	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv", ".ruff_cache"}})

	return afterChanges.Changes(r.Python.Base.Directory("/app"))

}

// Runs ruff check and returns the results in a json file.
func (r *Ruff) Report() *dagger.File {
	// Run ruff check with the provided output format
	return r.Python.Container().
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(
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

// Runs ruff format and returns a container that will fail on any errors.
func (r *Ruff) Format(
	// file pattern to exclude from ruff format
	// +optional
	exclude []string) *dagger.Container {
	args := []string{
		"uv",
		"run",
		"--with=ruff",
		"ruff",
		"format",
		".",
		"--diff",
		"--exclude=.venv/", //hack needed to get around ruff bug overriding default excludes
	}

	// exclude any given file patterns
	for _, exclude := range exclude {
		args = append(args, "--exclude", exclude)
	}

	return r.Python.Container().
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args)

}

// Runs ruff format and attempts to fix any format errors. Returns a Changeset
// that can be used to apply any changes found
// to the host.
func (r *Ruff) FormatFix(
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
		"--exclude=.venv/", //hack needed to get around ruff bug overriding default excludes
	}

	// exclude any given file patterns
	for _, exclude := range exclude {
		args = append(args, "--exclude", exclude)
	}
	ctr := r.Python.Container().
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args)

	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv/", ".ruff_cache/"}})
	return afterChanges.Changes(r.Python.Base.Directory("/app"))

}
