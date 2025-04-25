// Govulncheck reports known vulnerabilities in dependencies.
package main

import (
	"dagger/govulncheck/internal/dagger"
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
) *Govulncheck {
	if Container == nil {
		Container = defaultContainer(Version)
	} else {
		Container = Container.WithExec([]string{"go", "install", fmt.Sprintf("%s@%s", goVulnCheck, Version)})
	}

	return &Govulncheck{
		Container: Container,
		Flags:     []string{"govulncheck"},
	}
}

// Mount netrc credentials for a private git repository.
func (gv *Govulncheck) WithNetrc(
	// NETRC credentials
	netrc *dagger.Secret,
) *Govulncheck {
	gv.Container = gv.Container.WithMountedSecret("/root/.netrc", netrc)
	return gv
}

// Run govulncheck with a source directory.
//
// e.g. `govulncheck -mode=source`.
func (gv *Govulncheck) RunWithSource(
	// Go source directory
	source *dagger.Directory,
	// file patterns to include,
	// +optional
	// +default="./..."
	patterns string,
) *dagger.Container {
	gv.Flags = append(gv.Flags, patterns)
	return gv.Container.WithWorkdir("/work/src").
		WithMountedDirectory("/work/src", source).
		WithExec(gv.Flags)
}

// Run govulncheck with a binary.
//
// e.g. `govulncheck -mode=binary <binary>`.
func (gv *Govulncheck) RunWithBinary(
	// binary file
	binary *dagger.File,
) *dagger.Container {
	binaryPath := "/work/binary"
	gv.Container = gv.Container.WithMountedFile(binaryPath, binary)

	// perhaps unnecessary, but matches the usage docs in `govulncheck --help`
	args := append([]string{"-mode=binary"}, gv.Flags...)
	args = append(args, binaryPath)

	return gv.Container.WithExec(args)
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
