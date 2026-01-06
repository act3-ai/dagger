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

// Runs pytest and returns a container that will fail on any errors.
func (pt *Pytest) Test(
	// unit test directory
	// +optional
	// +default="test"
	unitTestDir string,
	// extra arguments to pytest, e.g., add "--cov-fail-under=80" to fail if coverage is below 80%
	// +optional
	extraArgs []string,
) *dagger.Container {

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
	}
	args = append(args, extraArgs...)

	return pt.Python.Container().
		WithExec(args)

}

// Runs pytest and returns results in multiple file formats. Current formats: json, xml, and html.
func (pt *Pytest) Report(
	// unit test directory
	// +optional
	// +default="test"
	unitTestDir string,
) (*PytestResults, error) {

	ctr := pt.Python.Container().
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
			}, dagger.ContainerWithExecOpts{
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
