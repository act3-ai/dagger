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

func New(
	// Custom container to use as a base container. Must have 'yamllint' available on PATH.
	// +optional
	Container *dagger.Container,

	// Version of yamllint to use, defaults to latest version available to apk.
	// +optional
	// +default="latest"
	Version string,
) *Yamllint {
	if Container == nil {
		Container = defaultContainer(Version)
	}

	return &Yamllint{
		Container: Container,
		Flags:     []string{"yamllint"},
	}
}

// Run 'yamllint' with all previously provided options.
//
// May be used as a "catch-all" in case functions are not implemented.
func (y *Yamllint) Run(ctx context.Context,
	// directory containing, but not limited to, YAML files to be linted.
	src *dagger.Directory,
	// flags, without 'yamllint'
	// +optional
	extraFlags []string,
) *dagger.Container {
	y.Flags = append(y.Flags, extraFlags...)

	// we could support a set of files, in addition to a directory, but
	// having a singular required arg avoids usage errors (optional dir or
	// set of files)
	srcPath := "src"
	y.Container = y.Container.WithMountedDirectory(srcPath, src).WithWorkdir(srcPath)
	y.Flags = append(y.Flags, ".")

	return y.Container.WithExec(y.Flags)
}

// List YAML files that can be linted.
//
// e.g. 'yamllint --list-files'.
func (y *Yamllint) ListFiles(ctx context.Context) ([]string, error) {
	y.Flags = append(y.Flags, "--list-files")
	out, err := y.Container.WithExec(y.Flags).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing yaml files: %w", err)
	}
	return strings.Split(out, "\n"), nil
}

// Mount a custom configuration file.
//
// e.g. 'yamllint --config-file <config>'.
func (y *Yamllint) WithConfig(
	// configuration file
	config *dagger.File,
) *Yamllint {
	cfgPath := ".yamllint.yaml"
	y.Container = y.Container.WithMountedFile(cfgPath, config)
	y.Flags = append(y.Flags, "--config-file", cfgPath)
	return y
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
