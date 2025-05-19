// Markdownlint provides utilities for running markdownlint-cli2 without installing locally with npm, brew, or docker. See https://github.com/DavidAnson/markdownlint-cli2 for more info.

package main

import (
	"context"
	"dagger/markdownlint/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "docker.io/davidanson/markdownlint-cli2"

type Markdownlint struct {
	Base *dagger.Container

	// +private
	Args []string
}

func New(
	// Custom container to use as a base container. Must have 'markdownlint-cli2' available on PATH.
	// +optional
	base *dagger.Container,

	// Version (image tag) to use as a markdownlint-cli2 binary source.
	// +optional
	// +default="latest"
	version string,
) *Markdownlint {
	if base == nil {
		base = dag.Container().
			From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	return &Markdownlint{
		Base: base,
		Args: []string{"markdownlint-cli2"},
	}
}

// Run markdownlint-cli2. Use the dagger native stdout to get the output, or export if the WithFix option was used.
func (m *Markdownlint) Run(ctx context.Context,
	// Source directory containing markdown files to be linted.
	// +ignore=["**", "!**/*.md", "!.markdownlint*", "!package.json"]
	src *dagger.Directory,

	// Additional arguments to pass to markdownlint-cli2, without 'markdownlint-cli2' itself.
	// +optional
	extraArgs []string,
) *dagger.Container {
	args := m.Args
	args = append(args, extraArgs...)
	return m.Base.
		WithWorkdir("/work/src").
		WithMountedDirectory(".", src).
		WithExec(args)
}

// WithFix updates files to resolve fixable issues (can be overriden in configuration).
//
// e.g. 'markdownlint-cli2 --fix'.
func (m *Markdownlint) WithFix() *Markdownlint {
	m.Args = append(m.Args, "--fix")
	return m
}

// Specify a custom configuration file.
//
// e.g. 'markdownlint-cli2 --config <config>'.
func (m *Markdownlint) WithConfig(
	// Custom configuration file
	config *dagger.File,
) *Markdownlint {
	// we cannot assume the file extension, and resolving it is fruitless
	cfgPath := ".markdownlint-cli2"
	m.Base = m.Base.WithMountedFile(cfgPath, config)
	m.Args = append(m.Args, "--config", cfgPath)
	return m
}
