// Govulncheck reports known vulnerabilities in dependencies.
//
// This module aids to run govulncheck anywhere, without managing it as a pipeline
// dependency.
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
	Command []string
}

func New(
	// Custom container to use as a base container. Must have go available. It's recommended to use github.com/sagikazarmark/daggerverse/go for a custom container, excluding the source directory.
	// +optional
	container *dagger.Container,

	// Version of govulncheck to use as a binary source.
	// +optional
	// +default="latest"
	version string,
) *Govulncheck {
	if container == nil {
		container = defaultContainer(version)
	} else {
		container = container.WithExec([]string{"go", "install", fmt.Sprintf("%s@%s", goVulnCheck, version)})
	}

	return &Govulncheck{
		Container: container,
		Command:   []string{"govulncheck"},
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
	cmd := gv.Command
	cmd = append(cmd, patterns)
	out, err := gv.Container.WithWorkdir(srcPath).
		WithMountedDirectory(srcPath, source).
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
	cmd := append([]string{"-mode=binary"}, gv.Command...)
	cmd = append(cmd, binaryPath)
	out, err := gv.Container.WithMountedFile(binaryPath, binary).
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

func (gv *Govulncheck) WithGitAuth(
	// registry's hostname
	address string,
	// username in registry
	username string,
	// password or token for registry
	secret *dagger.Secret,
) *Govulncheck {
	user := dag.SetSecret("username", username)
	netrc := dag.Netrc().WithLogin(address, user, secret)
	gv.Container = gv.Container.WithMountedSecret("/root/.netrc", netrc.AsSecret()).
		WithEnvVariable("GOPRIVATE", address)
	return gv
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
	gv.Command = append(gv.Command, "-db", url)
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
	gv.Command = append(gv.Command, "-format", format)
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
	gv.Command = append(gv.Command, "-scan", level)
	return gv
}

// Enable display of additional information.
//
// e.g. `govulncheck -show <enable>...`.
func (gv *Govulncheck) WithShow(
	// Enable additional info. Supported values: 'traces', 'color', 'version', and 'verbose'.
	enable []string,
) *Govulncheck {
	gv.Command = append(gv.Command, "-show", strings.Join(enable, ","))
	return gv
}

func defaultContainer(version string) *dagger.Container {
	return dag.Go().
		Exec([]string{"go", "install", fmt.Sprintf("%s@%s", goVulnCheck, version)})
}
