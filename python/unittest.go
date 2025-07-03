package main

import (
	"context"
	"dagger/python/internal/dagger"
)

// Return the result of running unit test
// returns directory "results/" with xml, junit, and html reports
func (python *Python) UnitTest(ctx context.Context,
	// unit test directory
	// +optional
	// +default="test"
	unitTestDir string,
) (*dagger.Directory, error) {

	c := python.Container().
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=pytest",
				"--with=pytest-cov",
				"pytest",
				unitTestDir,
				"--cov=.",
				"--cov-report",
				"term",
				"--cov-report",
				"xml:./results/unit-test.xml",
				"--cov-report",
				"html:./results/html/",
				"--junitxml=./results/pytest-junit.xml",
			})

	// Return a directory of test results in various formats
	return c.Directory("./results"), nil
}
