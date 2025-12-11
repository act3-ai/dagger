// A generated module for Tests functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/tests/internal/dagger"
	"fmt"
)

type Tests struct {
	// +private
	Source *dagger.Directory
}

func New(
	// testdata source directory
	// +defaultPath="testdata"
	src *dagger.Directory,
) *Tests {
	return &Tests{
		Source: src,
	}
}

// +check
// Test build for all platforms defined in goreleaser config.
func (t *Tests) BuildAll(ctx context.Context) error {
	dist := dag.Goreleaser(t.gitRepoGo()).
		Build().
		All()

	// goreleaser nests builds into subdirs, flatten and we'll multiply the expected builds
	// by 2 to account for subdirs
	distFlat, err := dist.Glob(ctx, "**/*")
	if err != nil {
		return err
	}

	expected := ((3 * 2) * 2) + 3 // [(len(goos) * len(goarch)) * 2 for subdirs] + 3 goreleaser files
	if len(distFlat) != expected {
		return fmt.Errorf("number of build results did not match build matrix, want %d, got %d", expected, len(distFlat))
	}

	return nil
}

// +check
// Test build for a specific platform.
func (t *Tests) BuildPlatform(ctx context.Context) error {
	bin := dag.Goreleaser(t.Source).
		Build().
		Platform("hello-world-linux-amd64", dagger.GoreleaserBuildPlatformOpts{Platform: dagger.Platform("linux/amd64")})
	if bin == nil {
		return fmt.Errorf("got nil build executable")
	}

	return nil
}

// gitRepoGo loads the go subset of testdata, turning it into a git repository.
// goreleaser requires a git repo for many of its functions.
func (t *Tests) gitRepoGo() *dagger.Directory {
	dir := "hello-world-go"

	return dag.Wolfi().
		Container(dagger.WolfiContainerOpts{Packages: []string{"git"}}).
		WithMountedDirectory(dir, t.Source.Directory(dir)).
		WithWorkdir(dir).
		WithExec([]string{"git", "init"}).
		WithExec([]string{"git", "config", "user.name", "test"}).
		WithExec([]string{"git", "config", "user.email", "test@dagger.io"}).
		// trick goreleaser into thinking we actually have a remote
		WithExec([]string{"git", "remote", "add", "origin", "git@github.com:foo/bar.git"}).
		WithExec([]string{"git", "add", "--all"}).
		WithExec([]string{"git", "commit", "-m", "fix: Initial commit"}).
		WithExec([]string{"git", "tag", "-a", "-m", "release(v0.2.0)", "v0.2.0"}).
		Directory(".")
}
