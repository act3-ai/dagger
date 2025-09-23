// Yamllint provides utility to lint YAML files without needing to download locally with pip or homebrew.
// It provides nearly all functionality given by yamllint, only exluding stdin uses.
//  See https://github.com/adrienverge/yamllint for more information.

package main

import (
	"context"
	"dagger/yamllint/internal/dagger"
	"errors"
	"fmt"
	"slices"
	"strings"
)

type Yamllint struct {
	// +private
	Base *dagger.Container

	// +private
	Command []string
}

func New(ctx context.Context,
	// Source directory containing markdown files to be linted.
	// +ignore=["**", "!**/*.yaml", "!**/*.yml", "!**/.yamllint*", "!**/.yamlignore*", "!**/.gitignore"]
	src *dagger.Directory,

	// Custom container to use as a base container. Must have 'yamllint' available on PATH.
	// +optional
	base *dagger.Container,

	// Version of yamllint to use, defaults to latest version available to apk.
	// +optional
	// +default="latest"
	version string,

	// Configuration file.
	// +optional
	config *dagger.File,
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

	args := []string{"yamllint"}
	srcDir := "/work/src"
	base = base.With(
		func(c *dagger.Container) *dagger.Container {
			if config != nil {
				cfgPath, err := config.Name(ctx)
				if err != nil {
					panic(fmt.Errorf("resolving configuration file name: %w", err))
				}
				c = c.WithMountedFile(cfgPath, config)
				args = append(args, "--config", cfgPath)
			}
			return c
		}).
		WithWorkdir(srcDir).
		WithMountedDirectory(srcDir, src)

	return &Yamllint{
		Base:    base,
		Command: []string{"yamllint"},
	}
}

// Run 'yamllint' with all previously provided options.
//
// May be used as a "catch-all" in case functions are not implemented.
func (y *Yamllint) Run(ctx context.Context,
	// Output results, without an error.
	// +optional
	ignoreError bool,

	// Output format. Supported values: 'parsable',' standard', 'colored', 'github', or 'auto'.
	// +optional
	// +default="auto"
	format string,

	// Additional arguments to pass to yamllint, without 'yamllint' itself.
	// +optional
	extraArgs []string,
) (string, error) {
	cmd := y.Command
	cmd = append(cmd, extraArgs...)
	cmd = append(cmd, "--format", format, ".")

	out, err := y.Base.
		WithExec(cmd).
		Stdout(ctx)

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

// List YAML files that can be linted.
//
// e.g. 'yamllint --list-files'.
func (y *Yamllint) ListFiles(ctx context.Context) ([]string, error) {
	cmd := y.Command
	cmd = append(cmd, "--list-files", ".")
	out, err := y.Base.WithExec(cmd).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing yaml files: %w", err)
	}

	files := strings.Split(out, "\n")
	return slices.DeleteFunc(files, func(s string) bool {
		return s == ""
	}), nil
}

// Return non-zero exit code on warnings as well as errors.
//
// e.g. 'yamllint --strict'.
func (y *Yamllint) WithStrict() *Yamllint {
	y.Command = append(y.Command, "--strict")
	return y
}

// Output only error level problems.
//
// e.g. 'yamllint --no-warnings'.
func (y *Yamllint) WithNoWarnings() *Yamllint {
	y.Command = append(y.Command, "--no-warnings")
	return y
}
