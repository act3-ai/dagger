package main

import "dagger/python/internal/dagger"

// run commands with flake8
type CognitiveComplexity struct {
	// +private
	Python *Python
}

// contains commands for running flake8 on a Python project.
func (p *Python) CognitiveComplexity() *CognitiveComplexity {
	return &CognitiveComplexity{Python: p}
}

// Runs a cognitive complexity lint using flake8
func (f *CognitiveComplexity) Lint(
	// file paths in source directory to scan
	// +optional
	// +default="."
	path string,
	// file paths to exclude
	// +optional
	exclude string,
) *dagger.Container {

	args := []string{
		"uvx",
		"--with=flake8",
		"--with=flake8-cognitive-complexity",
		"flake8",
		"--max-cognitive-complexity=15",
		"--select=CCR001",
	}

	if exclude != "" {
		args = append(args, "--exclude="+exclude)
	}

	args = append(args, path)

	return f.Python.Base.WithExec(args)

}
