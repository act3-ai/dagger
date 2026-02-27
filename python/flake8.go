package main

import "dagger/python/internal/dagger"

// run commands with flake8
type Flake8 struct {
	// +private
	Python *Python
}

// contains commands for running flake8 on a Python project.
func (p *Python) Flake8() *Flake8 {
	return &Flake8{Python: p}
}

// Runs Flake8 with flake8-cognitive-complexity
func (f *Flake8) Lint(
	// file path in source directory to scan
	// +optional
	// +default="src"
	path string,
) *dagger.Container {

	ctr := f.Python.Container().
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=flake8",
				"--with=flake8-cognitive-complexity",
				"flake8",
				"--max-cognitive-complexity=15",
				"--select=CCR001",
				path},
		)
	return ctr

}
