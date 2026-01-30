package main

import (
	"context"
	"dagger/python/internal/dagger"
	"strings"
)

type Ruff struct {
	// +private
	Python *Python
}

// contains commands for running ruff on a Python project.
// Will attempt to use ruff version specified in given project pyproject.toml, otherwise uses latest
func (p *Python) Ruff() *Ruff {
	return &Ruff{Python: p}
}

// Gets the ruff version from the dependency tree if it exists
func (r *Ruff) RuffVersion(ctx context.Context) string {
	// Use uv to get the version string for ruff and parse.
	ruffVersion, _ := r.Python.Base.WithExec([]string{"uv", "tree", "--frozen", "--package", "ruff"}).Stdout(ctx)

	if ruffVersion != "" {
		ruffVersion = strings.ReplaceAll(strings.Split(ruffVersion, "v")[1], "\n", "")
	}
	return ruffVersion
}

// Runs ruff check and returns a container that will fail on any errors.
func (r *Ruff) Lint(
	ctx context.Context,
	// +optional
	// +default="full"
	outputFormat string,
) *dagger.Container {
	ruffVersion := r.RuffVersion(ctx)

	withArg := "ruff"
	if ruffVersion != "" {
		withArg += "==" + ruffVersion // e.g., "ruff==0.1.1"
	}
	// Use the base image to avoid installing packages
	return r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=" + withArg,
				"--no-project",
				"ruff",
				"check", ".",
				"--output-format", outputFormat})

}

// Runs ruff check and attempts to fix any lint errors. Returns a changeset
// that can be used to apply any changes found
// to the host. Will return an error if any errors found are not considered fixable by ruff.
func (r *Ruff) LintFix(
	ctx context.Context,
	// +optional
	// +default="full"
	outputFormat string,
) *dagger.Changeset {
	ruffVersion := r.RuffVersion(ctx)

	withArg := "ruff"
	if ruffVersion != "" {
		withArg += "==" + ruffVersion // e.g., "ruff==0.1.1"
	}

	// Use the base image to avoid installing packages
	ctr := r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=" + withArg,
				"--no-project",
				"ruff",
				"check", ".",
				"--output-format", outputFormat,
				"--fix"})

	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv", ".ruff_cache"}})

	return afterChanges.Changes(r.Python.Base.Directory("/app"))

}

// Runs ruff check and returns the results in a json file.
func (r *Ruff) LintReport(ctx context.Context) *dagger.File {
	ruffVersion := r.RuffVersion(ctx)

	withArg := "ruff"
	if ruffVersion != "" {
		withArg += "==" + ruffVersion // e.g., "ruff==0.1.1"
	}
	// Use the base image to avoid installing packages
	return r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=" + withArg,
				"--no-project",
				"ruff",
				"check", ".",
				"--output-format",
				"json",
				"--output-file",
				"ruff-lint-results.json"},
			dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}).
		File("ruff-lint-results.json")

}

// Runs ruff format and returns a container that will fail on any errors.
func (r *Ruff) Format(
	ctx context.Context,
	// file pattern to exclude from ruff format
	// +optional
	exclude []string) *dagger.Container {
	ruffVersion := r.RuffVersion(ctx)

	withArg := "ruff"
	if ruffVersion != "" {
		withArg += "==" + ruffVersion // e.g., "ruff==0.1.1"
	}

	args := []string{
		"uv",
		"run",
		"--with=" + withArg,
		"--no-project",
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

	// Use the base image to avoid installing packages
	return r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args)

}

// Runs ruff format check and returns the results in a file.
func (r *Ruff) FormatReport(ctx context.Context,
	// file pattern to exclude from ruff format
	// +optional
	exclude []string) *dagger.File {
	ruffVersion := r.RuffVersion(ctx)

	withArg := "ruff"
	if ruffVersion != "" {
		withArg += "==" + ruffVersion // e.g., "ruff==0.1.1"
	}

	args := []string{
		"uv",
		"run",
		"--with=" + withArg,
		"--no-project",
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
	// Use the base image to avoid installing packages
	return r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny, RedirectStdout: "ruff-format-results.json"}).
		File("ruff-format-results.json")

}

// Runs ruff format and attempts to fix any format errors. Returns a Changeset
// that can be used to apply any changes found
// to the host.
func (r *Ruff) FormatFix(
	ctx context.Context,
	// file pattern to exclude from ruff format
	// +optional
	exclude []string) *dagger.Changeset {
	ruffVersion := r.RuffVersion(ctx)

	withArg := "ruff"
	if ruffVersion != "" {
		withArg += "==" + ruffVersion // e.g., "ruff==0.1.1"
	}

	args := []string{
		"uv",
		"run",
		"--with=" + withArg,
		"--no-project",
		"ruff",
		"format",
		".",
		"--exclude=.venv/", //hack needed to get around ruff bug overriding default excludes
	}

	// exclude any given file patterns
	for _, exclude := range exclude {
		args = append(args, "--exclude", exclude)
	}
	// Use the base image to avoid installing packages
	ctr := r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args)

	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv/", ".ruff_cache/"}})
	return afterChanges.Changes(r.Python.Base.Directory("/app"))

}
