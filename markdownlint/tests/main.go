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
	"strings"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// Run all tests
func (t *Tests) All(ctx context.Context,
	// +defaultPath="testdata"
	src *dagger.Directory) error {
	p := pool.New().WithErrors().WithContext(ctx)
	errDir := dag.Directory().WithFile("err.md", src.File("err.md")).WithFile(".markdownlint-cli2.yaml", src.File(".markdownlint-cli2.yaml"))
	validDir := dag.Directory().WithFile("valid.md", src.File("valid.md")).WithFile(".markdownlint-cli2.yaml", src.File(".markdownlint-cli2.yaml"))

	p.Go(func(ctx context.Context) error {
		return t.Run(ctx, validDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.RunWithErr(ctx, errDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.IgnoreErr(ctx, errDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.AutoFix(ctx, errDir)
	})

	return p.Wait()
}

// Returns results of run with valid markdown
func (t *Tests) Run(ctx context.Context,
	// +defaultPath="testdata"
	// +ignore=["err.md"]
	src *dagger.Directory) error {

	_, err := dag.Markdownlint(src).Run(ctx)

	return err
}

// Returns results of run with expected error
func (t *Tests) RunWithErr(ctx context.Context,
	// +defaultPath="testdata"
	// +ignore=["valid.md"]
	src *dagger.Directory) error {

	_, err := dag.Markdownlint(src).Run(ctx)
	if err != nil && strings.Contains(err.Error(), "MD047") {
		return nil
	}
	return err
}

// Returns results of run with err, but expected pass since error is being ignored
func (t *Tests) IgnoreErr(ctx context.Context,
	// +defaultPath="testdata"
	// +ignore=["valid.md"]
	src *dagger.Directory) error {

	_, err := dag.Markdownlint(src).Run(ctx, dagger.MarkdownlintRunOpts{IgnoreError: true})
	if err != nil {
		return fmt.Errorf("failed to ignore exec errors: %w", err)
	}
	return nil
}

// Returns results of run with err, but expected pass since error is being ignored
func (t *Tests) AutoFix(ctx context.Context,
	// +defaultPath="testdata"
	// +ignore=["valid.md"]
	src *dagger.Directory) error {

	_, err := dag.Markdownlint(src).AutoFix().Sync(ctx)
	// if err != nil {
	// 	return fmt.Errorf("failed to ignore exec errors: %w", err)
	// }
	return err
}
