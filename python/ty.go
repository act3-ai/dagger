package main

import (
	"dagger/python/internal/dagger"
)

type Ty struct {
	// +private
	Python *Python
}

// contains commands for running pyright on a Python project.
func (p *Python) Ty() *Ty {
	return &Ty{Python: p}
}

// Runs pyright on a given source directory. Returns a container that will fail on any errors.
func (t *Ty) Lint() *dagger.Container {

	return t.Python.Project().
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=ty",
				"ty",
				"check",
				".",
			})

}

// Runs pyright and returns results in a json file.
func (t *Ty) Report() *dagger.File {

	return t.Python.Project().
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=ty",
				"ty",
				"check",
				"--output-format",
				"gitlab",
				".",
			},
			dagger.ContainerWithExecOpts{
				Expect:         dagger.ReturnTypeAny,
				RedirectStdout: "ty-results.json"}).
		File("ty-results.json")

}
