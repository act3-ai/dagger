package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
)

type UnitTestResults struct {
	// prints the results to stdout
	Stdout string
	// returns results of unit-test as xml in a file.
	Xml *dagger.File
	// returns results of unit-test as json in a file.
	Json *dagger.File
	// returns results of unit-test as html in a directory
	Html *dagger.Directory
	// returns exit code of unit-test
	ExitCode int
	// A directory with all results merged in
	Merged *dagger.Directory
}

// Runs pytest and returns results in multiple formats.
// Current formats: Stdout, json, xml, and html.
func (python *Python) UnitTest(ctx context.Context,
	// unit test directory
	// +optional
	// +default="test"
	unitTestDir string,
) (*UnitTestResults, error) {

	ctr, err := python.Container().
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
				"xml:/results.xml",
				"--cov-report",
				"html:/html/",
				"--cov-report",
				"json:/results.json",
			}).Sync(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running unit-test: %w", err)
	}
	out, err := ctr.Stdout(ctx)
	if err != nil {
		// error getting stdout
		return nil, fmt.Errorf("get stdout code: %w", err)
	}
	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	xml := ctr.File("/results.xml")

	json := ctr.File("/results.json")

	html := ctr.Directory("/html")

	//merge all result files into a single directory
	merged := dag.Directory().WithFile("results.xml", xml).
		WithFile("results.json", json).
		WithDirectory("html", html)

	return &UnitTestResults{
		Stdout:   out,
		Xml:      xml,
		Json:     json,
		Html:     html,
		ExitCode: exitCode,
		Merged:   merged,
	}, nil
}
