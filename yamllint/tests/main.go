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
	errFile         = "err.yaml"
	validFile       = "valid.yaml"
	configFile      = ".yamllint.yml"      // basic config, used for most tests
	configFileWarn  = ".yamllint-warn.yml" // config used for warning level tests
	configFileNoExt = ".yamllint-foo"      // no extension, a yamllint convention with custom suffix
	gitIgnore       = ".gitignore"         // common yamlignore file
	yamlIgnore      = ".yamlignore"        // common yamlignore file
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

// Run all tests.
func (t *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx).WithMaxGoroutines(5)

	// Options for 'New'
	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "DirectoryFilter")
		defer span.End()
		return t.DirectoryFilter(ctx)
	})
	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "Config")
		defer span.End()
		return t.Config(ctx)
	})
	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "Version")
		defer span.End()
		return t.Version(ctx)
	})
	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "Base")
		defer span.End()
		return t.Base(ctx)
	})

	// Options for 'Run'
	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "IgnoreErr")
		defer span.End()
		return t.IgnoreErr(ctx)
	})
	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "OutputFormat")
		defer span.End()
		return t.OutputFormat(ctx)
	})
	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "ExtraArgs")
		defer span.End()
		return t.ExtraArgs(ctx)
	})

	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "ListFiles")
		defer span.End()
		return t.ListFiles(ctx)
	})

	// Modifiers for 'Run'
	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "WithStrict")
		defer span.End()
		return t.WithStrict(ctx)
	})
	p.Go(func(ctx context.Context) error {
		ctx, span := Tracer().Start(ctx, "WithNoWarnings")
		defer span.End()
		return t.WithNoWarnings(ctx)
	})

	return p.Wait()
}

// DirectoryFilter ensures the pre-call filtering is properly setup.
func (t *Tests) DirectoryFilter(ctx context.Context) error {
	testDir := dag.Directory().
		WithFile(configFile, t.Source.File(configFile)).           // .yml
		WithFile(validFile, t.Source.File(validFile)).             // .yaml
		WithFile(configFileNoExt, t.Source.File(configFileNoExt)). // .yamllint
		WithFile(gitIgnore, t.Source.File(gitIgnore)).             // .gitignore
		WithFile(yamlIgnore, t.Source.File(yamlIgnore))            // .yamlignore

	entries, err := testDir.Entries(ctx)
	if err != nil {
		return fmt.Errorf("getting number of expected entries: %w", err)
	}
	expectedEntries := len(entries)

	// can't access container directly, so try to use the config file AFTER filter.
	// specifying the config via constructor options would invalidate this test.
	out, err := dag.Yamllint(testDir).
		Run(ctx, dagger.YamllintRunOpts{ExtraArgs: []string{"-c", configFileNoExt, "--list-files"}})
	if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}

	gotEntries := strings.Fields(out)
	if len(gotEntries) != expectedEntries {
		return fmt.Errorf("unexpected number of entries, expected %d, got %d: values %v", expectedEntries, len(gotEntries), gotEntries)
	}

	return nil
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
