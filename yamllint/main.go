// Yamllint provides utility to lint YAML files without needing to download locally with pip or homebrew. It provides nearly all functionality given by yamllint, only exluding stdin uses. See https://github.com/adrienverge/yamllint for more information.

package main

import (
	"context"
	"dagger/yamllint/internal/dagger"
	"fmt"
	"strings"
)

type Yamllint struct {
	Base *dagger.Container

	// +private
	Args []string
}

func New(
	// Custom container to use as a base container. Must have 'yamllint' available on PATH.
	// +optional
	base *dagger.Container,

	// Version of yamllint to use, defaults to latest version available to apk.
	// +optional
	// +default="latest"
	version string,
) *Yamllint {
	if base == nil {
		// https://pkgs.alpinelinux.org/package/edge/community/x86_64/yamllint
		pkg := "yamllint"
		if version != "latest" {
			pkg = fmt.Sprintf("%s=%s", pkg, version)
		}
		base = dag.Wolfi().
			Container(
				dagger.WolfiContainerOpts{
					Packages: []string{pkg},
				},
			)
	}

	return &Yamllint{
		Base: base,
		Args: []string{"yamllint"},
	}
}

// Run 'yamllint' with all previously provided options.
//
// May be used as a "catch-all" in case functions are not implemented.
func (y *Yamllint) Run(ctx context.Context,
	// directory containing, but not limited to, YAML files to be linted.
	// +ignore=["**", "!**/*.yaml", "!**/*.yml"]
	src *dagger.Directory,
	// extra command line arguments
	// +optional
	extraArgs []string,
) *dagger.Container {
	args := y.Args
	args = append(args, extraArgs...)
	args = append(args, ".")

	srcPath := "src"
	return y.Base.WithMountedDirectory(srcPath, src).
		WithWorkdir(srcPath).
		WithExec(args)
}

// List YAML files that can be linted.
//
// e.g. 'yamllint --list-files'.
func (y *Yamllint) ListFiles(ctx context.Context) ([]string, error) {
	args := y.Args
	args = append(args, "--list-files")
	out, err := y.Base.WithExec(args).
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
	y.Base = y.Base.WithMountedFile(cfgPath, config)
	y.Args = append(y.Args, "--config-file", cfgPath)
	return y
}

// Specify output format.
//
// e.g. 'yamllint --format <format>'.
func (y *Yamllint) WithFormat(
	// output format. Supported values: 'parsable',' standard', 'colored', 'github', or 'auto'.
	format string,
) *Yamllint {
	y.Args = append(y.Args, "--format", format)
	return y
}

// Return non-zero exit code on warnings as well as errors.
//
// e.g. 'yamllint --strict'.
func (y *Yamllint) WithStrict() *Yamllint {
	y.Args = append(y.Args, "--strict")
	return y
}

// Output only error level problems.
//
// e.g. 'yamllint --no-warnings'.
func (y *Yamllint) WithNoWarnings() *Yamllint {
	y.Args = append(y.Args, "--no-warnings")
	return y
}
