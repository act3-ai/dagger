package main

import (
	"dagger/goreleaser/internal/dagger"
	"strconv"
	"strings"

	"github.com/containerd/platforms"
)

// Release represents the `goreleaser build` command.
type Build struct {
	// +private
	Goreleaser *Goreleaser

	// build command
	// +private
	Command []string
}

// Build represents the `goreleaser build` command.
func (gr *Goreleaser) Build() *Build {
	return &Build{
		Goreleaser: gr,
		Command:    []string{"goreleaser", "build"},
	}
}

// Build for a specific platform.
//
// e.g. `goreleaser build --single-target` with $GOOS, $GOARCH, and $GOARM set appropriately.
func (b *Build) Platform(
	// output file name
	outFile string,
	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	// +optional
	// +default="linux/amd64"
	platform dagger.Platform,
) *dagger.File {
	p := platforms.MustParse(string(platform))

	cmd := b.Command
	cmd = append(cmd, "--single-target", "--output", outFile)

	return b.Goreleaser.Container.
		WithEnvVariable(envGOOS, p.OS).
		WithEnvVariable(envGOARCH, p.Architecture).
		With(func(c *dagger.Container) *dagger.Container {
			if p.Variant != "" {
				return c.WithEnvVariable(envGOARM, p.Variant)
			}

			return c
		}).
		WithExec(cmd).
		File(outFile)
}

// Build for all platforms, defined in .goreleaser.yaml. Returns the goreleaser 'dist' directory.
//
// e.g. `goreleaser build`.
func (b *Build) All() *dagger.Directory {
	// TODO: Ideally, we would only return the executables. But that requires parsing the goreleaser
	// config for the target platforms.
	return b.Goreleaser.Container.
		WithExec(b.Command).
		Directory("dist")
}

// Build an unversioned snapshot, skipping all validations.
//
// e.g. `goreleaser build --snapshot`.
func (b *Build) WithSnapshot() *Build {
	b.Command = append(b.Command, "--snapshot")
	return b
}

// Automatically sets WithSnapshot if the repository is dirty.
//
// e.g. `goreleaser build --auto-snapshot`.
func (b *Build) WithAutoSnapshot() *Build {
	b.Command = append(b.Command, "--auto-snapshot")
	return b
}

// Removes the 'dist' directory before building.
//
// e.g. `goreleaser build --clean`.
func (b *Build) WithClean() *Build {
	b.Command = append(b.Command, "--clean")
	return b
}

// Timeout to the entire build process.
//
// e.g. `goreleaser build --timeout <duration>`.
func (b *Build) WithTimeout(
	// Timeout duration, e.g. 10m, 10m30s. Default is 30m.
	duration string,
) *Build {
	b.Command = append(b.Command, "--timeout", duration)
	return b
}

// Skip options: before, pre-hooks, post-hooks, validate.
//
// e.g. `goreleaser build --skip before,pre-hooks,...`.
func (b *Build) WithOptionSkip(
	// Skip options
	skip []string,
) *Build {
	b.Command = append(b.Command, "--skip", strings.Join(skip, ","))
	return b
}

// Tasks to run concurrently (default: number of CPUs).
//
// e.g. `goreleaser build --parallelism <n>`.
func (b *Build) WithParallelism(
	// concurrent tasks
	n int,
) *Build {
	b.Command = append(b.Command, "parallelism", strconv.Itoa(n))
	return b
}

// TODO: ensure this builds the flag correctly
// Builds only the specified build ids, as defined in a goreleaser configuration file.
//
// e.g. `goreleaser build --id <id> <id> ...`
// func (b *Build) WithIDs(
// 	// Build IDs
// 	ids []string,
// ) *Build {
// 	b.flags = append(b.flags, "--id")
// 	b.flags = append(b.flags, ids...)
// 	return b
// }
