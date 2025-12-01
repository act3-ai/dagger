// A generated module for Tests functions
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
	"dagger/tests/internal/dagger"
	"fmt"

	"github.com/dagger/dagger/util/parallel"
)

type Tests struct{}

// Run all tests.
func (t *Tests) All(ctx context.Context) error {
	return parallel.New().WithLimit(3).
		WithJob("Check coverage", t.Check).
		WithJob("Check coverage with excludes", t.CheckExcludes).
		WithJob("HTML", t.HTML).
		WithJob("SVG", t.SVG).
		WithJob("Summary", t.Summary).
		WithJob("Exec", t.Summary).
		// WithJob("Merge", t.Merge).
		Run(ctx)
}

func (t *Tests) base() *dagger.Container {
	src := dag.CurrentModule().Source().Directory("testdata/hello-world-go")
	return dag.
		Go().
		WithSource(src).
		WithCgoDisabled().
		Container()
}

var opts = dagger.CoverageOpts{Excludes: []string{`.gen.go`}}

func (t *Tests) Debug(ctx context.Context) (string, error) {
	results := dag.Coverage(t.base(), opts).UnitTests()
	return results.Summary().Contents(ctx)
}

// Test unit test coverage check
func (t *Tests) Check(ctx context.Context) error {
	results := dag.Coverage(t.base()).UnitTests()
	return results.Check(ctx, 71)
}

// Test unit test coverage check
func (t *Tests) CheckExcludes(ctx context.Context) error {
	results := dag.Coverage(t.base(), opts).UnitTests()
	return results.Check(ctx, 80)
}

// Test HTML generation
func (t *Tests) HTML(ctx context.Context) error {
	results := dag.Coverage(t.base()).UnitTests()
	html := results.HTML()
	i, err := html.Size(ctx)
	if err != nil {
		return err
	}
	if i < 100 {
		return fmt.Errorf("expected a larger HTML document: %d B", i)
	}
	return nil
}

// Test SVG generation
func (t *Tests) SVG(ctx context.Context) error {
	results := dag.Coverage(t.base()).UnitTests()
	svg := results.Svg()
	i, err := svg.Size(ctx)
	if err != nil {
		return err
	}
	if i < 100 {
		return fmt.Errorf("expected a larger SVG document: %d B", i)
	}
	return nil
}

// Test Summary generation
func (t *Tests) Summary(ctx context.Context) error {
	results := dag.Coverage(t.base()).UnitTests()
	summary := results.Summary()
	i, err := summary.Size(ctx)
	if err != nil {
		return err
	}
	if i < 100 {
		return fmt.Errorf("expected a larger Summary document: %d B", i)
	}
	return nil
}

// Test exec
func (t *Tests) Exec(ctx context.Context) error {
	results := dag.Coverage(t.base()).Exec("./cmd/myapp", []string{"argument"})

	return results.Check(ctx, 19)
}

/*
// Test merge
func (t *Tests) Merge(ctx context.Context) error {
	results1 := dag.Coverage(t.base(), opts).UnitTests()
	results2 := dag.Coverage(t.base()).Exec("./cmd/myapp", []string{"argument"})

	results := results1.Merge(results2)

	cov, err := results.Percent(ctx)
	if err != nil {
		return err
	}

	expected := 100.0
	if math.Abs(cov-expected) > 0.01 {
		return fmt.Errorf("expected %0.2f%% code coverage but saw %0.2f%%", expected, cov)
	}
	return nil
}
*/
