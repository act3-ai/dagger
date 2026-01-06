// Testing module for yamllint.
package main

import (
	"context"
	"dagger/tests/internal/dagger"
	"fmt"
	"strings"
)

// testdata dir entries
const (
	configFixErr = "cfgs/.yamllint-fixerr.yml" // fixes all errors in err.yaml
)

type Tests struct{}

// +check
// DirectoryFilter ensures the pre-call filtering is properly setup.
func (t *Tests) DirectoryFilter(ctx context.Context) error {
	src := t.srcDir()
	cfg := src.File(configFixErr)
	expectedEntries := 1 // if the config is used, then the ignore files should also be used, leaving us only with `valid.yaml`

	// also test for .yamllint* directive
	entries, err := dag.Yamllint(src, dagger.YamllintOpts{Config: cfg}).
		ListFiles(ctx)
	if err != nil {
		return err
	}

	if len(entries) != expectedEntries {
		return fmt.Errorf("unexpected number of entries, expected %d, got %d: values %v", expectedEntries, len(entries), entries)
	}

	return nil
}

// +check
// Config validates config mounting by mounting a yml file with errors and a config that ignores them
func (t *Tests) Config(ctx context.Context) *dagger.Container {
	src := t.srcDir()
	cfg := src.File(configFixErr)

	return dag.Yamllint(src, dagger.YamllintOpts{Config: cfg}).Lint()
}

// +check
// Version tests the 'Version' option for 'New'.
func (t *Tests) Version(ctx context.Context) error {
	version := "1.36.0"
	out, err := dag.Yamllint(t.srcDir(), dagger.YamllintOpts{Version: version}).
		Lint(dagger.YamllintLintOpts{ExtraArgs: []string{"--version"}}).
		Stdout(ctx)
	if err != nil {
		return err
	}

	if !strings.Contains(out, version) {
		return fmt.Errorf("unexpected version, want %s, got %s", version, out)
	}
	return nil
}

// +check
// Base tests the 'Base' option for 'New'.
func (t *Tests) Base(ctx context.Context) error {
	version := "1.36.0"
	base := dag.Wolfi().
		Container(dagger.WolfiContainerOpts{Packages: []string{fmt.Sprintf("yamllint=%s", version)}})

	out, err := dag.Yamllint(t.srcDir(), dagger.YamllintOpts{Base: base}).
		Lint(dagger.YamllintLintOpts{ExtraArgs: []string{"--version"}}).
		Stdout(ctx)
	if err != nil {
		return err
	}

	if !strings.Contains(out, version) {
		return fmt.Errorf("unexpected version, want %s, got %s", version, out)
	}
	return nil
}

// +check
// OutputFormat tests the 'Format' option for 'yamllint'.
func (t *Tests) OutputFormat(ctx context.Context) error {
	out, err := dag.Yamllint(t.srcDir()).
		Report(dagger.YamllintReportOpts{Format: "github"}).
		Contents(ctx)
	if err != nil && strings.Contains(out, "::group::") {
		return nil
	}
	return err
}

// +check
// ListFiles tests the 'ListFiles' method.
func (t *Tests) ListFiles(ctx context.Context) error {
	files, err := dag.Yamllint(t.srcDir()).ListFiles(ctx)
	switch {
	case err != nil:
		return err
	case len(files) != 2:
		return fmt.Errorf("unexpected number of files, want %d, got %d: %v", 2, len(files), files)
	default:
		return nil
	}
}

// +check
// WithStrict tests the 'WithStrict' method.
func (t *Tests) WithStrict(ctx context.Context) error {
	src := t.srcDir()
	cfg := src.File(configFixErr)
	// expect err due to 'document-start' warning
	report := dag.Yamllint(t.srcDir(), dagger.YamllintOpts{Config: cfg}).
		WithStrict().
		Report()

	expectedErr, err := report.Contents(ctx)
	switch {
	case err != nil:
		return fmt.Errorf("unexpected error: %w", err)
	case !strings.Contains(expectedErr, "document-start"):
		return fmt.Errorf("unexpected output, with no warnings test case not properly setup: output = %w", expectedErr)
	default:
		return nil
	}
}

// +check
// WithNoWarnings tests the 'WithNoWarnings' method.
func (t *Tests) WithNoWarnings(ctx context.Context) *dagger.Container {
	src := t.srcDir()
	cfg := src.File(configFixErr)
	// expect no err, despite a 'document-start' warning
	return dag.Yamllint(t.srcDir(), dagger.YamllintOpts{Config: cfg}).
		WithNoWarnings().
		Lint()

}

// validSrc returns a dir with a yamllint config and a test file that expects no issues.
func (t *Tests) srcDir() *dagger.Directory {
	return dag.CurrentModule().
		Source().
		Directory("testdata")

}
