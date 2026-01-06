package main

import (
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
	// +optional
	// +default="text"
	outputFormat string,
) *dagger.Container {

	ctr := pl.Python.Container().
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
		)
	return ctr

}

// Runs pylint and returns results in a json file
func (pl *Pylint) Report() *dagger.File {

	return pl.Python.Container().
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
		File("pylint-results.json")

}
