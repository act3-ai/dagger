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

	"golang.org/x/sync/errgroup"
)

type Tests struct {
	// +private
	Source *dagger.Directory
}

func New(
	// testdata source directory
	// +defaultPath="testdata"
	src *dagger.Directory,
) *Tests {
	return &Tests{
		Source: src,
	}
}

// Run all tests.
func (t *Tests) All(ctx context.Context) error {
	p, ctx := errgroup.WithContext(ctx)

	p.Go(func() error { return t.Check(ctx) })
	p.Go(func() error { return t.HTML(ctx) })
	p.Go(func() error { return t.SVG(ctx) })
	p.Go(func() error { return t.Summary(ctx) })

	return p.Wait()
}

func (t *Tests) base() *dagger.Container {
	src := dag.CurrentModule().Source().Directory("testdata/hello-world-go")
	return dag.
		Go().
		WithSource(src).
		WithCgoDisabled().
		Container()
}

// Test unit test coverage check
func (t *Tests) Check(ctx context.Context) error {
	results := dag.Coverage(t.base()).UnitTests()
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
