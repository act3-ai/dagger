package main

import (
	"context"
	"dagger/python/internal/dagger"
)

// run commands with mypy
type Mypy struct {
	// +private
	Python *Python
}

// contains commands for running mypy on a Python project.
func (p *Python) Mypy() *Mypy {
	return &Mypy{Python: p}
}

// Runs mypy on a given source directory. Returns a container that will fail on any errors.
func (m *Mypy) Lint(
	ctx context.Context,
	// +optional
	outputFormat string) (*dagger.Container, error) {
	args := []string{
		"uv",
		"run",
		"--with=mypy",
		"mypy",
	}

	// Append outputFormat only if it's provided
	if outputFormat != "" {
		args = append(args, "--output", outputFormat)
	}

	// Add path
	args = append(args, ".")
	ctr, err := m.Python.Runtime(ctx)
	if err != nil {
		return nil, err
	}

	return ctr.
		WithExec(args), nil

}

// Runs mypy and returns results in a json file.
func (m *Mypy) Report(ctx context.Context) (*dagger.File, error) {
	ctr, err := m.Python.Runtime(ctx)
	if err != nil {
		return nil, err
	}

	return ctr.
		WithExec([]string{
			"uv",
			"run",
			"--with=mypy",
			"mypy",
			"--output",
			"json",
			"."},
			dagger.ContainerWithExecOpts{
				Expect:         dagger.ReturnTypeAny,
				RedirectStdout: "mypy-results.json"}).
		File("mypy-results.json"), nil

}
