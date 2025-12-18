// Yamllint is a utility that lints YAML files without needing to download locally with pip or homebrew.

// It provides nearly all functionality given by yamllint by accepting a source directory or file.
//  See https://github.com/adrienverge/yamllint for more information.

package main

import (
	"context"
	"dagger/yamllint/internal/dagger"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type Yamllint struct {
	// +private
	Base *dagger.Container

	// +private
	Command []string
}

type YamllintResults struct {
	// returns results of yamllint as a file
	Results *dagger.File
	// returns exit code of yamllint
	// +private
	ExitCode int
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
				args = append(args, "--config-file", filepath.Join("/", cfgPath))
			}
			return c
		}).
		WithWorkdir(srcDir).
		WithMountedDirectory(srcDir, src)

	return &Yamllint{
		Base:    base,
		Command: args,
	}
}

// Runs 'yamllint' with all previously provided 'with' options.
func (y *Yamllint) Lint(ctx context.Context,
	// Output format. Supported values: 'parsable',' standard', 'colored', 'github', or 'auto'.
	// +optional
	// +default="auto"
	format string,

	// Additional arguments to pass to yamllint, without 'yamllint' itself.
	// +optional
	extraArgs []string,
) (*YamllintResults, error) {
	cmd := y.Command
	cmd = append(cmd, extraArgs...)
	cmd = append(cmd, "--format", format, ".")

	ctr, err := y.Base.
		WithExec(cmd, dagger.ContainerWithExecOpts{
			Expect: dagger.ReturnTypeAny}).Sync(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running yamllint: %w", err)
	}

	output, err := ctr.CombinedOutput(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("getting results: %w", err)
	}

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	return &YamllintResults{
		Results:  dag.File("yamllint-results.txt", output),
		ExitCode: exitCode,
	}, nil

}

// Check for any errors running yamllint
func (yr *YamllintResults) Check(ctx context.Context,
) error {
	if yr.ExitCode == 0 {
		return nil
	}
	results, err := yr.Results.Contents(ctx)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", results)
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
