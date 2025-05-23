// Govulncheck reports known vulnerabilities in dependencies.
package main

import (
	"context"
	"dagger/govulncheck/internal/dagger"
	"errors"
	"fmt"
	"strings"
)

// TODO: Support -mode=archive ??

const (
	goVulnCheck = "golang.org/x/vuln/cmd/govulncheck" // default: "latest"

	imageGo = "golang:latest" // github.com/sagikazarmark/daggerverse/go convention
)

type Govulncheck struct {
	Container *dagger.Container

	// +private
	Flags []string
}

func New(
	// Custom container to use as a base container. Must have go available. It's recommended to use github.com/sagikazarmark/daggerverse/go for a custom container, excluding the source directory.
	// +optional
	Container *dagger.Container,

	// Version of govulncheck to use as a binary source.
	// +optional
	// +default="latest"
	Version string,

	// Mount netrc credentials for a private git repository.
	// +optional
	Netrc *dagger.Secret,
) *Govulncheck {
	if Container == nil {
		Container = defaultContainer(Version)
	} else {
		Container = Container.WithExec([]string{"go", "install", fmt.Sprintf("%s@%s", goVulnCheck, Version)})
	}

	Container = Container.With(
		func(c *dagger.Container) *dagger.Container {
			if Netrc != nil {
				c = c.WithMountedSecret("/root/.netrc", Netrc)
			}
			return c
		})

	return &Govulncheck{
		Container: Container,
		Flags:     []string{"govulncheck"},
	}
}

// Run govulncheck on a source directory.
//
// e.g. `govulncheck -mode=source`.
func (gv *Govulncheck) ScanSource(ctx context.Context,
	// Go source directory
	// +ignore=["**", "!**/go.mod", "!**/go.sum", "!**/*.go"]
	source *dagger.Directory,
	// Output results, without an error.
	// +optional
	ignoreError bool,
	// file patterns to include,
	// +optional
	// +default="./..."
	patterns string,
) (string, error) {
	srcPath := "/work/src"
	gv.Flags = append(gv.Flags, patterns)
	out, err := gv.Container.WithWorkdir(srcPath).
		WithMountedDirectory(srcPath, source).
		WithExec(gv.Flags).
		Stdout(ctx)

	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		result := fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr)
		if ignoreError {
			return result, nil
		}
		// linter exit code != 0
		return "", fmt.Errorf("%s", result)
	case err != nil:
		// some other dagger error, e.g. graphql
		return "", err
	default:
		// stdout of the linter with exit code 0
		return out, nil
	}
}

// Run govulncheck on a binary.
//
// e.g. `govulncheck -mode=binary <binary>`.
func (gv *Govulncheck) ScanBinary(ctx context.Context,
	// binary file
	binary *dagger.File,
	// Output results, without an error.
	// +optional
	ignoreError bool,
) (string, error) {
	binaryPath := "/work/binary"
	args := append([]string{"-mode=binary"}, gv.Flags...)
	args = append(args, binaryPath)
	out, err := gv.Container.WithMountedFile(binaryPath, binary).
		WithExec(args).
		Stdout(ctx)

	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		result := fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr)
		if ignoreError {
			return result, nil
		}
		// linter exit code != 0
		return "", fmt.Errorf("%s", result)
	case err != nil:
		// some other dagger error, e.g. graphql
		return "", err
	default:
		// stdout of the linter with exit code 0
		return out, nil
	}
}

// Specify a vulnerability database url.
//
// e.g. `govlulncheck -db <url>`.
func (gv *Govulncheck) WithDB(
	// vulnerability database url.
	// +optional
	// +default="https://vuln.go.dev"
	url string,
) *Govulncheck {
	gv.Flags = append(gv.Flags, "-db", url)
	return gv
}

// Specify the output format.
//
// e.g. `govulncheck -format <format>`.
func (gv *Govulncheck) WithFormat(
	// Output format. Supported values: 'text', 'json', 'sarif', and 'openvex'.
	// +optional
	// +default="text"
	format string,
) *Govulncheck {
	gv.Flags = append(gv.Flags, "-format", format)
	return gv
}

// Set the scanning level.
//
// e.g. `govulncheck -scan <level>`.
func (gv *Govulncheck) WithScanLevel(
	// scanning level. Supported values: 'module', 'package', or 'symbol'.
	// +optional
	// +default="symbol"
	level string,
) *Govulncheck {
	gv.Flags = append(gv.Flags, "-scan", level)
	return gv
}

// Enable display of additional information.
//
// e.g. `govulncheck -show <enable>...`.
func (gv *Govulncheck) WithShow(
	// Enable additional info. Supported values: 'traces', 'color', 'version', and 'verbose'.
	enable []string,
) *Govulncheck {
	gv.Flags = append(gv.Flags, "-show", strings.Join(enable, ","))
	return gv
}

func defaultContainer(version string) *dagger.Container {
	return dag.Go().
		Exec([]string{"go", "install", fmt.Sprintf("%s@%s", goVulnCheck, version)})
}
