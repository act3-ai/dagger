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

func New(ctx context.Context,
	// Source directory containing markdown files to be linted.
	Src *dagger.Directory,

	// Custom container to use as a base container. Must have 'markdownlint-cli2' available on PATH.
	// +optional
	Container *dagger.Container,

	// Version (image tag) to use as a markdownlint-cli2 binary source.
	// +optional
	// +default="latest"
	Version string,

	// '.markdownlint-cli2.*' configuration file.
	// +optional
	Config *dagger.File,
) *Markdownlint {
	if Container == nil {
		Container = defaultContainer(Version)
	}

	flags := []string{"markdownlint-cli2"}
	srcDir := "/work/src"
	Container = Container.With(
		func(c *dagger.Container) *dagger.Container {
			if Config != nil {
				cfgPath, err := Config.Name(ctx)
				if err != nil {
					panic(fmt.Errorf("resolving configuration file name: %w", err))
				}
				c = c.WithMountedFile(cfgPath, Config)
				flags = append(flags, "--config", cfgPath)
			}
			return c
		}).
		WithWorkdir(srcDir).
		WithMountedDirectory(srcDir, Src)

	return &Markdownlint{
		Container: Container,
		Flags:     flags,
	}
}

// Run markdownlint-cli2. Typical usage is to run to detect an error, and, if an
// error is returned, re-run with `--results` to return the output.
func (m *Markdownlint) Run(ctx context.Context,
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
		WithExec(m.Flags, dagger.ContainerWithExecOpts{Expect: expect}).
		Stdout(ctx)
}

// AutoFix updates files to resolve fixable issues (can be overriden in configuration).
//
// e.g. 'markdownlint-cli2 --fix'.
func (m *Markdownlint) AutoFix() *dagger.Directory {
	m.Flags = append(m.Flags, "--fix")
	return m.Container.
		WithExec(m.Flags).
		Directory("/work/src")
}

func defaultContainer(version string) *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
}
