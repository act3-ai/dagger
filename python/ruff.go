package main

import (
	"context"
	"dagger/python/internal/dagger"
	"slices"
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
func (r *Ruff) version(ctx context.Context) string {
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
	// extra arguments to pass to the ruff lint check
	// +optional
	extraArgs []string,
) *dagger.Container {
	ruffVersion := r.version(ctx)

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
		"check"}
	args = append(args, extraArgs...)
	// Use the base image to avoid installing packages
	return r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args)

}

// Runs ruff check and attempts to fix any lint errors. Returns a changeset
// that can be used to apply any changes found
// to the host. Will return an error if any errors found are not considered fixable by ruff.
func (r *Ruff) LintFix(
	ctx context.Context,
	// extra arguments to pass to the ruff lint check
	// +optional
	extraArgs []string,
) *dagger.Changeset {
	ruffVersion := r.version(ctx)

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
		"check",
		"--fix",
	}
	args = append(args, extraArgs...)
	// Use the base image to avoid installing packages
	ctr := r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args)

	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv", ".ruff_cache"}})

	return afterChanges.Changes(r.Python.Base.Directory("/app"))

}

// Runs ruff check and returns the results in a file.
func (r *Ruff) LintReport(
	ctx context.Context,
	// extra arguments to pass to the ruff lint check
	// +optional
	extraArgs []string,
) *dagger.File {
	ruffVersion := r.version(ctx)

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
		"check"}

	outputFileName := "ruff-lint-results.json"

	if !slices.Contains(extraArgs, "--output-file") {
		args = append(args, "--output-file", outputFileName)
	} else {
		outputFileName = extraArgs[slices.Index(extraArgs, "--output-file")+1]
	}
	args = append(args, extraArgs...)

	// Use the base image to avoid installing packages
	return r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}).
		File(outputFileName)

}

// Runs ruff format and returns a container that will fail on any errors.
func (r *Ruff) Format(
	ctx context.Context,
	// extra arguments to pass to the ruff format check
	// +optional
	extraArgs []string) *dagger.Container {
	ruffVersion := r.version(ctx)

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
	}
	// if the user already specifies diff/check don't specify it again
	if !slices.Contains(extraArgs, "--diff") && !slices.Contains(extraArgs, "--check") {
		args = append(args, "--diff")
	}
	args = append(args, extraArgs...)

	// Use the base image to avoid installing packages
	return r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args)

}

// Runs ruff format check and returns the results in a file. The results file
// is named 'ruff-format-results.json' regardless of any output format specified through
// extraArgs.
func (r *Ruff) FormatReport(
	ctx context.Context,
	// extra arguments to pass to the ruff format check
	// +optional
	extraArgs []string) *dagger.File {
	ruffVersion := r.version(ctx)

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
	}
	// if the user already specifies diff/check don't specify it again
	if !slices.Contains(extraArgs, "--diff") && !slices.Contains(extraArgs, "--check") {
		args = append(args, "--diff")
	}
	args = append(args, extraArgs...)

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
	// extra arguments to pass to the ruff format check
	// +optional
	extraArgs []string) *dagger.Changeset {
	ruffVersion := r.version(ctx)

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
	}
	args = append(args, extraArgs...)
	// Use the base image to avoid installing packages
	ctr := r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache")).
		WithExec(args)

	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv/", ".ruff_cache/"}})
	return afterChanges.Changes(r.Python.Base.Directory("/app"))

}
