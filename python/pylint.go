package main

import (
	"context"
	"dagger/python/internal/dagger"
)

type Pylint struct {
	// +private
	Python *Python
}

// contains commands for running pylint on a Python project.
func (p *Python) Pylint() *Pylint {
	return &Pylint{Python: p}
}

// Runs pylint on a given source directory. Returns a container that will fail on any errors.
func (pl *Pylint) Lint(
	ctx context.Context,
	// +optional
	// +default="text"
	outputFormat string,
) (*dagger.Container, error) {
	ctr, err := pl.Python.Runtime(ctx)
	if err != nil {
		return nil, err
	}

	return ctr.
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=pylint",
				"pylint",
				"--recursive=y",
				"--persistent=n",
				"--ignore-paths=.venv",
				"--output-format",
				outputFormat,
				"."},
		), nil

}

// Runs pylint and returns results in a json file
func (pl *Pylint) Report(ctx context.Context) (*dagger.File, error) {
	ctr, err := pl.Python.Runtime(ctx)
	if err != nil {
		return nil, err
	}

	return ctr.
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=pylint",
				"pylint",
				"--recursive=y",
				"--persistent=n",
				"--ignore-paths=.venv",
				"--output-format",
				"json2",
				"--output",
				"pylint-results.json",
				"."},
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny}).
		File("pylint-results.json"), nil

}
