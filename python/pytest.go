package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
)

type Pytest struct {
	// +private
	Python *Python
}
type PytestResults struct {
	// prints the combined output of stdout and stderr as a string
	// +private
	Output string
	// returns results of unit-test as xml in a file.
	Xml *dagger.File
	// returns results of unit-test as json in a file.
	Json *dagger.File
	// returns results of unit-test as html in a directory
	Html *dagger.Directory
	// returns exit code of unit-test
	// +private
	ExitCode int
	// A directory with all results merged in
	Merged *dagger.Directory
}

// Runs pytest and returns results in multiple file formats. Current formats: CombinedOutput, json, xml, and html.
func (p *Python) Pytest(ctx context.Context,
	// unit test directory
	// +optional
	// +default="test"
	unitTestDir string,

	// extra arguments to pytest, e.g., add "--cov-fail-under=80" to fail if coverage is below 80%
	// +optional
	extraArgs []string,
) (*PytestResults, error) {
	args := []string{
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
	}
	args = append(args, extraArgs...)

	ctr, err := p.Container().
		WithExec(args, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}).
		Sync(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running pytest: %w", err)
	}

	out, err := ctr.CombinedOutput(ctx)
	if err != nil {
		// error getting stdout
		return nil, fmt.Errorf("get combined output: %w", err)
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

	return &PytestResults{
		Output:   out,
		Xml:      xml,
		Json:     json,
		Html:     html,
		ExitCode: exitCode,
		Merged:   merged,
	}, nil
}

func (pt *PytestResults) Check() error {
	if pt.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("%s", pt.Output)
}
