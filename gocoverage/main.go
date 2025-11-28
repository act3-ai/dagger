// Package go code coverage module

package main

import (
	"context"
	"dagger/coverage/internal/dagger"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Code coverage generator
type Coverage struct {
	// +private
	Base *dagger.Container
}

func New(
	// base container for go.  Expectations:
	// - working directory is set to the go.mod file
	// - go tool works for go-cover-treemap (optional)
	base *dagger.Container,
) *Coverage {
	return &Coverage{
		Base: base,
	}
}

// Code coverage from unit tests
func (m *Coverage) UnitTests(
	ctx context.Context,
) (*CoverageResults, error) {
	pkg, err := m.Base.WithExec([]string{"go", "list", "-m"}).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the module name: %w: ", err)
	}
	pkg = strings.TrimSpace(pkg)

	// TODO produce binary coverage results instead
	// see https://github.com/thediveo/lxkns/blob/cef5a31d7517cb126378f81628f51672cb793527/scripts/cov.sh#L28
	// go test -cover ./... -args -test.gocoverdir /coverage
	raw := m.Base.
		WithExec([]string{"go", "test", "./...", "-coverprofile", "/raw", "-coverpkg", pkg + "/..."}).
		File("/raw")

	filter, err := m.Base.Exists(ctx, "filter-coverage.sh")
	if err != nil {
		return nil, err
	}
	if filter {
		raw = m.Base.
			WithFile("/raw", raw).
			WithExec([]string{"./filter-coverage.sh"}, dagger.ContainerWithExecOpts{
				RedirectStdin:  "/raw",
				RedirectStdout: "/filtered",
			}).
			File("/filtered")
	}

	return &CoverageResults{
		Base: m.Base,
		Raw:  raw,
	}, nil
}

/*
// Run a go package with coverage and combine the results
func (m *Coverage) Exec(ctx context.Context,
	pkg string,
	args []string,
) (*CoverageResults, error) {
	raw = m.Base.
		WithExec([]string{"go", "build", "-cover", pkg, "-o", "/app"}).
		WithEnvVariable("GOCOVERDIR", "/coverage").
		WithExec(append([]string{"/app"}, args...)).
		Directory("/coverage")

	return &CoverageResults{
		Base: m.Base,
		Raw:  raw,
	}, nil
}
*/

// Code coverage results
type CoverageResults struct {
	// +private
	Base *dagger.Container

	// Raw coverage results
	Raw *dagger.File
}

/*
func (m *CoverageResults) Merge(ctx context.Context,
	other []*CoverageResults) *CoverageResults {
	raw := m.Base.
		WithExec([]string{"go", "tool", "covdata", "merge", "-i", "-o", "/raw1,/raw2", "/raw"}).
		File("/raw")

}
*/

// SVG heat map of code coverage
func (m *CoverageResults) SVG() *dagger.File {
	return m.Base.
		WithFile("/raw", m.Raw).
		WithExec([]string{"go", "tool", "go-cover-treemap", "-coverprofile", "/raw"}, dagger.ContainerWithExecOpts{
			RedirectStdout: "/heat.svg",
		}).
		File("/heat.svg")
}

// HTML report
func (m *CoverageResults) HTML() *dagger.File {
	return m.Base.
		WithFile("/raw", m.Raw).
		WithExec([]string{"go", "tool", "cover", "-html", "/raw", "-o", "/index.html"}).
		File("/index.html")
}

// Summary of coverage by functions
func (m *CoverageResults) Summary() *dagger.File {
	return m.Base.
		WithFile("/raw", m.Raw).
		WithExec([]string{"go", "tool", "cover", "-func", "/raw", "-o", "/summary.txt"}).
		File("/summary.txt")
}

// Percent code coverage
func (m *CoverageResults) Percent(ctx context.Context) (float64, error) {
	// percent, err := m.Base.
	// 	WithFile("/raw", m.Raw).
	// 	WithExec([]string{"go", "tool", "covdata", "percent", "-i", "/raw"}).
	// 	Stdout(ctx)
	// if err != nil {
	// 	return math.NaN(), err
	// }

	summary, err := m.Summary().Contents(ctx)
	if err != nil {
		return math.NaN(), err
	}

	re := regexp.MustCompile(`total:\s+\(statements\)\s+(?<percentage>.*)%`)
	match := re.FindAllStringSubmatch(summary, -1)
	if len(match) != 1 && len(match[0]) != 1 {
		return math.NaN(), fmt.Errorf("expected a single match for %q in: %q", re, summary)
	}
	percentage, err := strconv.ParseFloat(match[0][1], 64)
	if err != nil {
		return math.NaN(), fmt.Errorf("percentage is not a float: %w", err)
	}

	return percentage, nil
}

// Check that the coverage percentage is above a threshold
func (m *CoverageResults) Check(ctx context.Context,
	// minimum percentage to accept
	threshold float64,
) error {
	percent, err := m.Percent(ctx)
	if err != nil {
		return err
	}
	if percent < threshold {
		return fmt.Errorf("Code coverage of %.2f%% is insufficient (need at least %.2f%%)", percent, threshold)
	}
	return nil
}

// Get a directory of all the results
func (m *CoverageResults) Directory(ctx context.Context) (*dagger.Directory, error) {
	percent, err := m.Percent(ctx)
	if err != nil {
		return nil, err
	}

	return dag.Directory().
			WithFile("heat", m.SVG()).
			WithFile("index.html", m.HTML()).
			WithNewFile("percent", strconv.FormatFloat(percent, 'f', 2, 64)),
		nil
}
