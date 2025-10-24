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
	"reflect"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// Run all tests.
func (t *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(t.Prepare)

	return p.Wait()
}

// return container with a git repo and an initial commit with tag v1.0.0
func (t *Tests) gitRepo() *dagger.Container {

	return dag.Container().
		From("alpine/git").
		WithWorkdir("/repo").
		WithNewFile("README.md", "# My Project").
		WithExec([]string{"git", "init"}).
		WithExec([]string{"git", "config", "user.name", "test"}).
		WithExec([]string{"git", "config", "user.email", "test@dagger.io"}).
		WithExec([]string{"git", "add", "README.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: Initial commit"}).
		WithExec([]string{"git", "tag", "-a", "-m", "Initial commit", "v1.0.0"})

}

// Run all tests
// func (t *Tests) All(ctx context.Context) error {
// 	p := pool.New().WithErrors().WithContext(ctx)

// 	return p.Wait()
// }

// test prepare
func (t *Tests) Prepare(ctx context.Context) error {

	gitref := t.gitRepo().
		WithNewFile("test.md", "test").
		WithExec([]string{"git", "add", "test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	expectedDir := dag.Directory().
		WithNewFile("VERSION", "1.0.1\n").
		WithNewFile("CHANGELOG.md", "test log").
		WithNewFile("releases/v1.0.1.md", "test release")

	actualDir := dag.Release(gitref).Prepare()

	expectedEntries, err := expectedDir.Entries(ctx)
	if err != nil {
		return err
	}

	actualEntries, err := actualDir.Entries(ctx)

	if !reflect.DeepEqual(expectedEntries, actualEntries) {
		return fmt.Errorf("files do not match:\nexpected: %v\ngot: %v", expectedEntries, actualEntries)
	}

	//hack needed because dagger does not search subdirectories when using entries
	releaseFileCheck, err := actualDir.Exists(ctx, "releases/v1.0.1.md")

	if !releaseFileCheck {
		return fmt.Errorf("files does not exist: releases/v1.0.1.md")
	}

	return err
}
