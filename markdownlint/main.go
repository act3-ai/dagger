// Markdownlint provides utilities for running markdownlint-cli2 without installing locally with npm, brew, or docker. See https://github.com/DavidAnson/markdownlint-cli2 for more info. Most configuration should be done in '.markdownlint-cli2.yaml' within the source directory, or provided using 'WithConfig'.

package main

import (
	"context"
	"dagger/markdownlint/internal/dagger"
	"errors"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "docker.io/davidanson/markdownlint-cli2"

type Markdownlint struct {
	Base *dagger.Container

	// +private
	Command []string

	// +private
	disableDefaultGlobs bool
}

func New(ctx context.Context,
	// Source directory containing markdown files to be linted.
	// +ignore=["**", "!**/*.md", "!.markdownlint*", "!package.json"]
	src *dagger.Directory,

	// Custom container to use as a base container. Must have 'markdownlint-cli2' available on PATH.
	// +optional
	base *dagger.Container,

	// Version (image tag) to use as a markdownlint-cli2 binary source.
	// +optional
	// +default="latest"
	version string,

	// Configuration file.
	// +optional
	config *dagger.File,
) *Markdownlint {
	if base == nil {
		base = dag.Container().
			From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	var disableDefaultGlobs bool
	cfgFiles, err := src.Filter(dagger.DirectoryFilterOpts{Include: []string{".markdownlint*"}}).
		Entries(ctx)
	if err != nil {
		panic(fmt.Errorf("discovering config files: %w", err))
	}
	if len(cfgFiles) > 0 {
		disableDefaultGlobs = true
	}

	cmd := []string{"markdownlint-cli2"}
	srcDir := "/work/src"
	base = base.With(
		func(c *dagger.Container) *dagger.Container {
			if config != nil {
				cfgPath, err := config.Name(ctx)
				if err != nil {
					panic(fmt.Errorf("resolving configuration file name: %w", err))
				}
				c = c.WithMountedFile(cfgPath, config)
				cmd = append(cmd, "--config", cfgPath)
				disableDefaultGlobs = true
			}
			return c
		}).
		WithWorkdir(srcDir).
		WithMountedDirectory(srcDir, src)

	return &Markdownlint{
		Base:                base,
		Command:             cmd,
		disableDefaultGlobs: disableDefaultGlobs,
	}
}

// Run markdownlint-cli2. Typical usage is to run to detect linting errors, and, if an
// error is returned, re-run with `--results` to return the output file or `--results contents`
// to output to stdout.
func (m *Markdownlint) Run(ctx context.Context,
	// Additional arguments to pass to markdownlint-cli2, without 'markdownlint-cli2' itself.
	// +optional
	extraArgs []string,

	// Output results, ignoring errors.
	// +optional
	ignoreError bool,
) (string, error) {
	cmd := m.Command
	cmd = append(cmd, extraArgs...)

	if !m.disableDefaultGlobs || len(extraArgs) <= 0 {
		// match all markdown files, see "Dot-only glob" https://github.com/DavidAnson/markdownlint-cli2?tab=readme-ov-file#command-line
		cmd = append(cmd, ".")
	}

	out, err := m.Base.WithExec(cmd).Stdout(ctx)
	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		// exit code != 0
		result := fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr)
		if ignoreError {
			return result, nil
		}
		return "", fmt.Errorf("%s", result)
	case err != nil:
		// some other dagger error, e.g. graphql
		return "", err
	default:
		// exit code 0
		return out, nil
	}
}

// AutoFix updates files to resolve fixable issues (can be overriden in configuration).
// It returns the entire source directory, use `export --path=<path-to-source>` to
// write the updates to the host.
//
// e.g. 'markdownlint-cli2 --fix'.
func (m *Markdownlint) AutoFix() *dagger.Directory {
	cmd := m.Command
	cmd = append(cmd, "--fix")
	return m.Base.
		WithExec(cmd).
		Directory("/work/src")
}
