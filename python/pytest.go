package main

import (
	"dagger/python/internal/dagger"
)

type Pytest struct {
	// +private
	Python *Python
}

// contains commands for running pytest on a Python project.
func (p *Python) Pytest() *Pytest {
	return &Pytest{Python: p}
}

type PytestResults struct {
	// returns results of unit-test as xml in a file.
	Xml *dagger.File
	// returns results of unit-test as json in a file.
	Json *dagger.File
	// returns results of unit-test as html in a directory
	Html *dagger.Directory
	// A directory with all results merged in
	Merged *dagger.Directory
}

// helper function for pytest args parsing
func (pt *Pytest) pytestArgs(
	testPaths []string,
	extraArgs []string,
) []string {

	args := []string{
		"uv",
		"run",
		"--with=pytest",
		"--with=pytest-cov",
		"pytest",
	}

	// Only restrict if paths were provided
	if len(testPaths) > 0 {
		args = append(args, testPaths...)
	}

	args = append(args,
		"--cov=.",
		"--cov-report", "term",
		"--cov-report", "xml:/results.xml",
		"--cov-report", "html:/html/",
		"--cov-report", "json:/results.json",
	)

	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}

	return args
}

// Runs pytest and returns a container that will fail on any errors.
func (pt *Pytest) Test(
	// provide optional test paths for pytest to use,
	// otherwise pytest will autodiscover from the given source dir
	// +optional
	testPaths []string,
	// extra arguments to pytest, e.g., add "--cov-fail-under=80" to fail if coverage is below 80%
	// +optional
	extraArgs []string,
) *dagger.Container {

	args := pt.pytestArgs(testPaths, extraArgs)

	return pt.Python.Container().
		WithExec(args)

}

// Runs pytest and returns results in multiple file formats. Current formats: json, xml, and html.
func (pt *Pytest) Report(
	// provide optional test paths for pytest to use,
	// otherwise pytest will autodiscover from the given source dir
	// +optional
	testPaths []string,
	// extra arguments to pytest, e.g., add "--cov-fail-under=80" to fail if coverage is below 80%
	// +optional
	extraArgs []string,
) (*PytestResults, error) {

	args := pt.pytestArgs(testPaths, extraArgs)

	ctr := pt.Python.Container().
		WithExec(
			args,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny})

	xml := ctr.File("/results.xml")

	json := ctr.File("/results.json")

	html := ctr.Directory("/html")

	//merge all result files into a single directory
	merged := dag.Directory().WithFile("results.xml", xml).
		WithFile("results.json", json).
		WithDirectory("html", html)

	return &PytestResults{
		Xml:    xml,
		Json:   json,
		Html:   html,
		Merged: merged,
	}, nil
}
