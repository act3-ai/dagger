// Yamllint provides utility to lint YAML files without needing to download locally with pip or homebrew. It provides nearly all functionality given by yamllint, only exluding stdin uses. See https://github.com/adrienverge/yamllint for more information.

package main

import (
	"context"
	"dagger/yamllint/internal/dagger"
	"fmt"
	"strings"
)

type Yamllint struct {
	Container *dagger.Container

	// +private
	Flags []string
}

func New(ctx context.Context,
	// Source directory containing markdown files to be linted.
	Src *dagger.Directory,

	// Custom container to use as a base container. Must have 'yamllint' available on PATH.
	// +optional
	Container *dagger.Container,

	// Version of yamllint to use, defaults to latest version available to apk.
	// +optional
	// +default="latest"
	Version string,

	// Configuration file.
	// +optional
	Config *dagger.File,
) *Yamllint {
	if Container == nil {
		Container = defaultContainer(Version)
	}

	flags := []string{"yamllint"}
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

	return &Yamllint{
		Container: Container,
		Flags:     []string{"yamllint"},
	}
}

// Run 'yamllint' with all previously provided options.
//
// May be used as a "catch-all" in case functions are not implemented.
func (y *Yamllint) Run(ctx context.Context,
	// Additional arguments to pass to yamllint, without 'yamllint' itself.
	// +optional
	extraArgs []string,

	// Output results, without an error.
	// +optional
	results bool,
) (string, error) {
	y.Flags = append(y.Flags, extraArgs...)
	y.Flags = append(y.Flags, ".")

	expect := dagger.ReturnTypeSuccess
	if results {
		expect = dagger.ReturnTypeAny
	}

	return y.Container.
		WithExec(y.Flags, dagger.ContainerWithExecOpts{Expect: expect}).
		Stdout(ctx)
}

// List YAML files that can be linted.
//
// e.g. 'yamllint --list-files'.
func (y *Yamllint) ListFiles(ctx context.Context) ([]string, error) {
	y.Flags = append(y.Flags, "--list-files", ".")
	out, err := y.Container.WithExec(y.Flags).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing yaml files: %w", err)
	}
	return strings.Split(out, "\n"), nil
}

// Specify output format.
//
// e.g. 'yamllint --format <format>'.
func (y *Yamllint) WithFormat(
	// output format. Supported values: 'parsable',' standard', 'colored', 'github', or 'auto'.
	format string,
) *Yamllint {
	y.Flags = append(y.Flags, "--format", format)
	return y
}

// Return non-zero exit code on warnings as well as errors.
//
// e.g. 'yamllint --strict'.
func (y *Yamllint) WithStrict() *Yamllint {
	y.Flags = append(y.Flags, "--strict")
	return y
}

// Output only error level problems.
//
// e.g. 'yamllint --no-warnings'.
func (y *Yamllint) WithNoWarnings() *Yamllint {
	y.Flags = append(y.Flags, "--no-warnings")
	return y
}

func defaultContainer(version string) *dagger.Container {
	// https://pkgs.alpinelinux.org/package/edge/community/x86_64/yamllint
	pkg := "yamllint"
	if version != "latest" {
		pkg = fmt.Sprintf("%s=%s", pkg, version)
	}
	return dag.Wolfi().
		Container(
			dagger.WolfiContainerOpts{
				Packages: []string{pkg},
			},
		)
}
