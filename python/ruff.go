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

func (r *Ruff) baseArgs(ctx context.Context, subcommand string) []string {
	ruffVersion := r.version(ctx)

	withArg := "ruff"
	if ruffVersion != "" {
		withArg += "==" + ruffVersion
	}

	return []string{
		"uv",
		"run",
		"--with=" + withArg,
		"--no-project",
		"ruff",
		subcommand,
	}
}

func (r *Ruff) baseContainer() *dagger.Container {
	return r.Python.Base.
		WithMountedCache("/app/.ruff_cache", dag.CacheVolume("ruff-cache"))
}

// Runs ruff check and returns a container that will fail on any errors.
func (r *Ruff) Lint(
	ctx context.Context,
	// extra arguments to pass to the ruff lint check
	// +optional
	extraArgs []string,
) *dagger.Container {
	args := r.baseArgs(ctx, "check")
	args = append(args, extraArgs...)

	return r.baseContainer().WithExec(args)
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

	args := r.baseArgs(ctx, "check")
	args = append(args, "--fix")
	args = append(args, extraArgs...)

	ctr := r.baseContainer().WithExec(args)
	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv", ".ruff_cache"}})

	return afterChanges.Changes(r.Python.Base.Directory("/app"))

}

// Runs ruff check and returns the results in a file.
func (r *Ruff) LintReport(
	ctx context.Context,
	// output format of the lint report
	// +optional
	// +default=json
	outputFormat string,
	// name of report file
	// +optional
	// +default=ruff-lint-results.json
	outputFile string,
) *dagger.File {
	args := r.baseArgs(ctx, "check")

	args = append(args, "--output-format", outputFormat, "--output-file", outputFile)

	return r.baseContainer().
		WithExec(args, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}).
		File(outputFile)
}

// Runs ruff format and returns a container that will fail on any errors.
func (r *Ruff) Format(
	ctx context.Context,
	// extra arguments to pass to the ruff format check
	// +optional
	extraArgs []string) *dagger.Container {
	args := r.baseArgs(ctx, "format")

	if !slices.Contains(extraArgs, "--diff") && !slices.Contains(extraArgs, "--check") {
		args = append(args, "--diff")
	}

	args = append(args, extraArgs...)
	return r.baseContainer().WithExec(args)

}

// Runs ruff format check and returns the results in a file.
func (r *Ruff) FormatReport(
	ctx context.Context,
	// output format of the lint report
	// +optional
	// +default=json
	outputFormat string,
	// name of report file
	// +optional
	// +default=ruff-format-results.json
	outputFile string) *dagger.File {
	args := r.baseArgs(ctx, "format")
	args = append(args, "--diff", "--output-format", outputFormat)

	return r.baseContainer().
		WithExec(args, dagger.ContainerWithExecOpts{
			Expect:         dagger.ReturnTypeAny,
			RedirectStdout: outputFile,
		}).
		File(outputFile)

}

// Runs ruff format and attempts to fix any format errors. Returns a Changeset
// that can be used to apply any changes found
// to the host.
func (r *Ruff) FormatFix(
	ctx context.Context,
	// extra arguments to pass to the ruff format check
	// +optional
	extraArgs []string) *dagger.Changeset {
	args := r.baseArgs(ctx, "format")
	args = append(args, extraArgs...)

	ctr := r.baseContainer().WithExec(args)

	afterChanges := ctr.Directory("/app").Filter(dagger.DirectoryFilterOpts{Exclude: []string{".venv/", ".ruff_cache/"}})
	return afterChanges.Changes(r.Python.Base.Directory("/app"))

}
