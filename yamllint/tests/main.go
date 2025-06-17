// Testing module for yamllint.
package main

import (
	"context"
	"dagger/tests/internal/dagger"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

// testdata dir entries
const (
	errFile        = "err.yaml"
	validFile      = "valid.yaml"
	configFile     = ".yamllint.yml"      // basic config, used for most tests
	configFileWarn = ".yamllint-warn.yml" // config used for warning level tests
)

type Tests struct {
	// +private
	Source *dagger.Directory
}

func New(
	// all testdata
	// +defaultPath="testdata"
	src *dagger.Directory,
) *Tests {
	return &Tests{
		Source: src,
	}
}

// Run all tests
func (t *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx).WithMaxGoroutines(5)

	// Options for 'New'
	p.Go(t.Config)
	p.Go(t.Version)
	p.Go(t.Base)
	// Options for 'Run'
	p.Go(t.IgnoreErr)
	p.Go(t.OutputFormat)
	p.Go(t.ExtraArgs)

	p.Go(t.ListFiles)

	// Modifiers for 'Run'
	p.Go(t.WithStrict)
	p.Go(t.WithNoWarnings)

	return p.Wait()
}

// Config validates config mounting.
func (t *Tests) Config(ctx context.Context) error {
	src := t.validSrc().WithoutFile(configFile)
	cfg := t.Source.File(configFile)

	_, err := dag.Yamllint(src, dagger.YamllintOpts{Config: cfg}).Run(ctx)
	switch {
	case err != nil && strings.Contains(err.Error(), "new-line-at-end-of-file"):
		// custom config defers from default by the 'new-line-at-end-of-file' rule
		// there's likely a bug with config mounting
		return fmt.Errorf("suspected default configuration used, config mouting error: %w", err)
	case err != nil:
		// a change in testdata/* has affected this test
		return fmt.Errorf("unexpected err, config mount test case not properly setup: %w", err)
	default:
		return nil
	}
}

// Version tests the 'Version' option for 'New'.
func (t *Tests) Version(ctx context.Context) error {
	version := "1.36.0"
	out, err := dag.Yamllint(t.errSrc(), dagger.YamllintOpts{Version: version}).Run(ctx, dagger.YamllintRunOpts{ExtraArgs: []string{"--version"}})
	if err != nil {
		return err
	}

	if !strings.Contains(out, version) {
		return fmt.Errorf("unexpected version, want %s, got %s", version, out)
	}
	return nil
}

// Base tests the 'Base' option for 'New'.
func (t *Tests) Base(ctx context.Context) error {
	version := "1.36.0"
	base := dag.Wolfi().
		Container(dagger.WolfiContainerOpts{Packages: []string{fmt.Sprintf("yamllint=%s", version)}})

	out, err := dag.Yamllint(t.validSrc(), dagger.YamllintOpts{Base: base}).
		Run(ctx, dagger.YamllintRunOpts{ExtraArgs: []string{"--version"}})
	switch {
	case err != nil:
		return err
	case !strings.Contains(out, version):
		return fmt.Errorf("expected version %s used in base container, got %s", version, out)
	default:
		return nil
	}
}

// IgnoreErr tests the 'IgnoreError' option for 'Run'.
func (t *Tests) IgnoreErr(ctx context.Context) error {
	_, err := dag.Yamllint(t.errSrc()).Run(ctx, dagger.YamllintRunOpts{IgnoreError: true})
	if err != nil {
		return fmt.Errorf("failed to ignore exec errors: %w", err)
	}
	return nil
}

// OutputFormat tests the 'Format' option for 'Run'.
func (t *Tests) OutputFormat(ctx context.Context) error {
	_, err := dag.Yamllint(t.errSrc()).Run(ctx, dagger.YamllintRunOpts{Format: "github"})
	if err != nil && strings.Contains(err.Error(), "::group::") {
		return nil
	}
	return err
}

// ExtraArgs tests the 'ExtraArgs' option for 'Run'.
func (t *Tests) ExtraArgs(ctx context.Context) error {
	// running with --version on err case should not return an err
	_, err := dag.Yamllint(t.errSrc()).Run(ctx, dagger.YamllintRunOpts{ExtraArgs: []string{"--version"}})
	return err
}

// ListFiles tests the 'ListFiles' method.
func (t *Tests) ListFiles(ctx context.Context) error {
	files, err := dag.Yamllint(t.validSrc()).ListFiles(ctx)
	switch {
	case err != nil:
		return err
	case len(files) != 2:
		return fmt.Errorf("unexpected number of files, want %d, got %d: %v", 2, len(files), files)
	default:
		return nil
	}
}

// WithStrict tests the 'WithStrict' method.
func (t *Tests) WithStrict(ctx context.Context) error {
	// expect err due to 'trailing-spaces' warning
	_, err := dag.Yamllint(t.validSrcWarn()).WithStrict().Run(ctx)
	switch {
	case err == nil:
		return fmt.Errorf("expected error on warnings, got nil error")
	case !strings.Contains(err.Error(), "trailing-spaces"):
		return fmt.Errorf("unexpected output, with no warnings test case not properly setup: output = %w", err)
	default:
		return nil
	}
}

// WithNoWarnings tests the 'WithNoWarnings' method.
func (t *Tests) WithNoWarnings(ctx context.Context) error {
	// expect no err, dispite a 'trailing-spaces' warning
	out, err := dag.Yamllint(t.validSrcWarn()).WithNoWarnings().Run(ctx)
	switch {
	case err != nil:
		return err
	case out != "":
		return fmt.Errorf("expected empty output, got %s", out)
	default:
		return nil
	}
}

// errSrc returns a dir with a yamllint config and a test file that should throw lint errs.
func (t *Tests) errSrc() *dagger.Directory {
	return dag.Directory().
		WithFile(configFile, t.Source.File(configFile)).
		WithFile(errFile, t.Source.File(errFile))

}

// validSrc returns a dir with a yamllint config and a test file that expects no issues.
func (t *Tests) validSrc() *dagger.Directory {
	return dag.Directory().
		WithFile(configFile, t.Source.File(configFile)).
		WithFile(validFile, t.Source.File(validFile))
}

// validSrcWarn returns a dir with a yamllint config with custom warning levels.
func (t *Tests) validSrcWarn() *dagger.Directory {
	return dag.Directory().
		// use standard config file name, so yamllint can discover it without extra work on our end
		WithFile(configFile, t.Source.File(configFileWarn)).
		WithFile(validFile, t.Source.File(validFile))
}
