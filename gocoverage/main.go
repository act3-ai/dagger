// A generated module for Coverage functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/coverage/internal/dagger"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Coverage struct{}

// Run unit test code coverage
func (m *Coverage) Run(
	ctx context.Context,
	// base container for go.  Expectations:
	// - working directory is set to the go.mod file
	// - go tool works for go-cover-treemap and cover
	base *dagger.Container,
) (*CoverageResults, error) {
	pkg, err := base.WithExec([]string{"go", "list", "-m"}).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the module name: %w: ", err)
	}
	pkg = strings.TrimSpace(pkg)

	d := base.
		WithDirectory("coverage", dag.Directory()).
		WithExec([]string{"go", "test", "./...", "-coverprofile", "coverage/raw", "-coverpkg", pkg + "/..."}).
		WithExec([]string{"./filter-coverage.sh"}, dagger.ContainerWithExecOpts{
			RedirectStdin:  "coverage/raw",
			RedirectStdout: "coverage/filtered",
		}).
		WithExec([]string{"go", "tool", "go-cover-treemap", "-coverprofile", "coverage/filtered"}, dagger.ContainerWithExecOpts{
			RedirectStdout: "coverage/heat.svg",
		}).
		WithExec([]string{"go", "tool", "cover", "-html", "coverage/filtered", "-o", "coverage/index.html"}).
		WithExec([]string{"go", "tool", "cover", "-func", "coverage/filtered", "-o", "coverage/summary.txt"}).
		Directory("coverage")

	summary, err := d.File("summary.txt").Contents(ctx)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`total:\s+\(statements\)\s+(?<percentage>.*)%`)
	match := re.FindAllStringSubmatch(summary, -1)
	if len(match) != 1 && len(match[0]) != 1 {
		return nil, fmt.Errorf("expected a single match for %q in: %q", re, summary)
	}
	percentage, err := strconv.ParseFloat(match[0][1], 64)
	if err != nil {
		return nil, fmt.Errorf("percentage is not a float: %w", err)
	}

	return &CoverageResults{
		SVG:      d.File("heat.svg"),
		HTML:     d.File("index.html"),
		Summary:  d.File("summary.txt"),
		Coverage: percentage,
	}, nil
}

type CoverageResults struct {
	// Heat map image
	SVG *dagger.File

	// HTML coverage report
	HTML *dagger.File

	// Summary by function
	Summary *dagger.File

	// Coverage percentage
	Coverage float64
}

// Check that the coverage percentage is above a threshold
func (m *CoverageResults) Check(threshold float64) error {
	if m.Coverage < threshold {
		return fmt.Errorf("Code coverage of %.2f is not above %.2f", m.Coverage, threshold)
	}
	return nil
}

// Get a directory of all the results
func (m *CoverageResults) Directory() *dagger.Directory {
	return dag.Directory().
		WithFiles("/", []*dagger.File{m.SVG, m.HTML, m.Summary}).
		WithNewFile("/percentage", fmt.Sprintf("%.2f", m.Coverage))
}
