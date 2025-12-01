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

	// +private
	Excludes []string
}

func New(
	// base container for go.  Expectations:
	// - working directory is set to the go.mod file
	// - go tool works for go-cover-treemap (optional)
	base *dagger.Container,

	// Exclude files from coverage
	// +optional
	excludes []string,
) *Coverage {
	return &Coverage{
		Base:     base,
		Excludes: excludes,
	}
}

// Code coverage from unit tests
func (m *Coverage) UnitTests(
	ctx context.Context,
) (*CoverageResults, error) {
	// produce binary coverage results instead of the traditional textual format
	// see https://github.com/thediveo/lxkns/blob/cef5a31d7517cb126378f81628f51672cb793527/scripts/cov.sh#L28

	/*
		// This does not work when there is no top level .go file.  Does not collect coverage properly.
		const covDir = "/coverage"
		raw := m.Base.
			WithDirectory(covDir, dag.Directory()).
			WithEnvVariable("GOCOVERDIR", covDir).
			WithExec([]string{"go", "test", "-cover", "-args", "-test.gocoverdir", covDir, "./..."}).
			Directory(covDir)
	*/

	// TODO use golang (don't shell out)
	pkg, err := m.Base.WithExec([]string{"go", "list", "-m"}).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the module name: %w: ", err)
	}
	pkg = strings.TrimSpace(pkg)

	text := m.Base.
		WithDirectory("coverage", dag.Directory()).
		WithExec([]string{"go", "test", "./...", "-coverprofile", "/raw.txt", "-coverpkg", pkg + "/..."}).
		File("/raw.txt")

	return &CoverageResults{
		Coverage: m,
		Text:     text,
	}, nil
}

// Run a go package with coverage
func (m *Coverage) Exec(ctx context.Context,
	pkg string,
	args []string,
) (*CoverageResults, error) {
	const covDir = "/coverage"
	raw := m.Base.
		WithDirectory(covDir, dag.Directory()).
		WithExec([]string{"go", "build", "-cover", "-o", "/app", pkg}).
		WithEnvVariable("GOCOVERDIR", covDir).
		WithExec(append([]string{"/app"}, args...)).
		Directory(covDir)

	return &CoverageResults{
		Coverage: m,
		Raw:      raw,
	}, nil
}

// Code coverage results
type CoverageResults struct {
	// +private
	Coverage *Coverage

	// Raw coverage results directory (binary format)
	// +private
	Raw *dagger.Directory

	// Coverage results in older format (text format)
	// +private
	Text *dagger.File
}

func (cr *CoverageResults) withRaw(raw *dagger.Directory) *CoverageResults {
	return &CoverageResults{
		Coverage: cr.Coverage,
		Raw:      raw,
	}
}

// Merge all results
func (cr *CoverageResults) Merge(ctx context.Context,
	// other coverage results to merge into the returned results
	other *CoverageResults,

	// module paths to include
	// +optional
	pkgs []string,
) (*CoverageResults, error) {
	if cr.Text != nil {
		return nil, fmt.Errorf("Coverage results has older format.  Merge is not supported.")
	}

	args := []string{"go", "tool", "covdata", "merge", "-i", "/left,/right", "-o", "/merged"}
	if len(pkgs) != 0 {
		args = append(args, "-pkg", strings.Join(pkgs, ","))
	}

	raw := cr.Coverage.Base.
		WithDirectory("/left", cr.Raw).
		WithDirectory("/right", other.Raw).
		WithDirectory("/merged", dag.Directory()).
		WithExec(args).
		Directory("/merged")

	return cr.withRaw(raw), nil
}

// Text format (older style) coverage format
func (cr *CoverageResults) TextFormat() *dagger.File {
	var cov *dagger.File = cr.Text

	// Do the conversion if needed
	if cov == nil {
		cov = cr.Coverage.Base.
			WithDirectory("/coverage", cr.Raw).
			WithExec([]string{"go", "tool", "covdata", "textfmt", "-i", "/coverage", "-o", "/coverage.txt"}).
			File("/coverage.txt")
	}

	// Filter if needed
	if len(cr.Coverage.Excludes) != 0 {
		args := []string{"grep", "-v"}
		for _, exclude := range cr.Coverage.Excludes {
			args = append(args, "-e", exclude)
		}
		cov = cr.Coverage.Base.
			WithFile("/coverage.txt", cov).
			WithExec(args,
				dagger.ContainerWithExecOpts{
					RedirectStdin:  "/coverage.txt",
					RedirectStdout: "/filtered.txt",
				}).
			File("/filtered.txt")
	}

	return cov
}

// SVG heat map of code coverage
func (cr *CoverageResults) SVG() *dagger.File {
	return cr.Coverage.Base.
		WithFile("/coverage.txt", cr.TextFormat()).
		WithExec([]string{"go", "tool", "go-cover-treemap", "-coverprofile", "/coverage.txt"}, dagger.ContainerWithExecOpts{
			RedirectStdout: "/heat.svg",
		}).
		File("/heat.svg")
}

// HTML report
func (cr *CoverageResults) HTML() *dagger.File {
	return cr.Coverage.Base.
		WithFile("/coverage.txt", cr.TextFormat()).
		WithExec([]string{"go", "tool", "cover", "-html", "/coverage.txt", "-o", "/index.html"}).
		File("/index.html")
}

// Summary of coverage by functions
func (cr *CoverageResults) Summary() *dagger.File {
	return cr.Coverage.Base.
		WithFile("/coverage.txt", cr.TextFormat()).
		WithExec([]string{"go", "tool", "cover", "-func", "/coverage.txt", "-o", "/summary.txt"}).
		File("/summary.txt")

	/*
		return m.Base.
			WithDirectory("/coverage", m.Raw).
			WithExec([]string{"go", "tool", "covdata", "func", "-i", "/coverage"}, dagger.ContainerWithExecOpts{
				RedirectStdout: "/summary.txt",
			}).
			File("/summary.txt")
	*/
}

// Percent code coverage (total of all statements)
func (cr *CoverageResults) Percent(ctx context.Context) (float64, error) {
	// percent, err := m.Base.
	// 	WithFile("/coverage", m.Raw).
	// 	WithExec([]string{"go", "tool", "covdata", "percent", "-i", "/raw"}).
	// 	Stdout(ctx)
	// if err != nil {
	// 	return math.NaN(), err
	// }

	summary, err := cr.Summary().Contents(ctx)
	if err != nil {
		return math.NaN(), err
	}

	re := regexp.MustCompile(`total:\s+\(statements\)\s+(?<percentage>.*)%`)
	match := re.FindAllStringSubmatch(summary, -1)
	if len(match) != 1 || len(match[0]) != 2 {
		return math.NaN(), fmt.Errorf("expected a single match for %s in: %s", re, summary)
	}
	percentage, err := strconv.ParseFloat(match[0][1], 64)
	if err != nil {
		return math.NaN(), fmt.Errorf("percentage is not a float: %w", err)
	}

	return percentage, nil
}

// Check that the coverage percentage is above a threshold
func (cr *CoverageResults) Check(ctx context.Context,
	// minimum percentage to accept
	threshold float64,
) error {
	percent, err := cr.Percent(ctx)
	if err != nil {
		return err
	}
	if percent < threshold {
		return fmt.Errorf("Code coverage of %.2f%% is insufficient (need at least %.2f%%)", percent, threshold)
	}
	return nil
}

// Get a directory of all the results
func (cr *CoverageResults) Directory(ctx context.Context) (*dagger.Directory, error) {
	percent, err := cr.Percent(ctx)
	if err != nil {
		return nil, err
	}

	return dag.Directory().
			WithFile("heat", cr.SVG()).
			WithFile("index.html", cr.HTML()).
			WithFile("coverage.txt", cr.TextFormat()).
			WithFile("summary.txt", cr.Summary()).
			WithNewFile("percent", strconv.FormatFloat(percent, 'f', 2, 64)),
		nil
}
