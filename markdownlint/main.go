// Markdownlint provides utilities for running markdownlint-cli2 without installing locally with npm, brew, or docker. See https://github.com/DavidAnson/markdownlint-cli2 for more info. Most configuration should be done in '.markdownlint-cli2.yaml' within the source directory, or provided using 'WithConfig'.

package main

import (
	"context"
	"dagger/markdownlint/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "docker.io/davidanson/markdownlint-cli2"

type Markdownlint struct {
	Container *dagger.Container

	// +private
	Flags []string
}

func New(
	// Custom container to use as a base container. Must have 'markdownlint-cli2' available on PATH.
	// +optional
	Container *dagger.Container,

	// Version (image tag) to use as a markdownlint-cli2 binary source.
	// +optional
	// +default="latest"
	Version string,
) *Markdownlint {
	if Container == nil {
		Container = defaultContainer(Version)
	}

	return &Markdownlint{
		Container: Container,
		Flags:     []string{"markdownlint-cli2"},
	}
}

// Run markdownlint-cli2. Typical usage is to run to detect an error, and, if an
// error is returned, re-run with `--results` to return the output.
func (m *Markdownlint) Run(ctx context.Context,
	// Source directory containing markdown files to be linted.
	src *dagger.Directory,

	// Additional arguments to pass to markdownlint-cli2, without 'markdownlint-cli2' itself.
	// +optional
	extraArgs []string,

	// Output results, without an error.
	// +optional
	results bool,
) (string, error) {
	m.Flags = append(m.Flags, extraArgs...)

	expect := dagger.ReturnTypeSuccess
	if results {
		expect = dagger.ReturnTypeFailure
	}

	return m.Container.
		WithWorkdir("/work/src").
		WithMountedDirectory(".", src).
		WithExec(m.Flags, dagger.ContainerWithExecOpts{Expect: expect}).
		Stdout(ctx)
}

// AutoFix updates files to resolve fixable issues (can be overriden in configuration).
//
// e.g. 'markdownlint-cli2 --fix'.
func (m *Markdownlint) AutoFix(
	// Source directory containing markdown files to be linted.
	src *dagger.Directory,
) *dagger.Directory {
	m.Flags = append(m.Flags, "--fix")
	return m.Container.
		WithWorkdir("/work/src").
		WithMountedDirectory(".", src).
		WithExec(m.Flags).
		Directory("/work/src")
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
	m.Container = m.Container.WithMountedFile(cfgPath, config)
	m.Flags = append(m.Flags, "--config", cfgPath)
	return m
}

func defaultContainer(version string) *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
}
