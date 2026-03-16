package main

import (
	"context"
	"dagger/python/internal/dagger"
)

type Pyright struct {
	// +private
	Python *Python
}

// contains commands for running pyright on a Python project.
func (p *Python) Pyright() *Pyright {
	return &Pyright{Python: p}
}

// Runs pyright on a given source directory. Returns a container that will fail on any errors.
func (pr *Pyright) Lint(ctx context.Context) (*dagger.Container, error) {
	ctr, err := pr.Python.Runtime(ctx)
	if err != nil {
		return nil, err
	}
	return ctr.
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=pyright",
				"pyright",
				".",
			}), nil

}

// Runs pyright and returns results in a json file.
func (pr *Pyright) Report(ctx context.Context) (*dagger.File, error) {
	ctr, err := pr.Python.Runtime(ctx)
	if err != nil {
		return nil, err
	}

	return ctr.
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=pyright",
				"pyright",
				".",
				"--outputjson",
			},
			dagger.ContainerWithExecOpts{
				Expect:         dagger.ReturnTypeAny,
				RedirectStdout: "pyright-results.json"}).
		File("pyright-results.json"), nil

}
