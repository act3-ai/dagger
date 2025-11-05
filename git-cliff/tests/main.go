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
	"errors"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// +private
func parseErr(err error) error {
	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		// exit code != 0
		return fmt.Errorf("%s", fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr))
	case err != nil:
		// some other dagger error, e.g. graphql
		return err
	default:
		// exit code 0
		return nil
	}
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
func (t *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(t.BumpedVersion)
	p.Go(t.BumpedVersionIncludePath)
	p.Go(t.BumpedVersionExcludePath)
	p.Go(t.WithBumpedVersion)
	p.Go(t.WithTagPattern)
	p.Go(t.WithLatest)
	p.Go(t.WithCurrent)
	p.Go(t.WithOutput)

	return p.Wait()
}

// test BumpedVersion
func (t *Tests) BumpedVersion(ctx context.Context) error {

	//version should be bumped with a fix: commit
	gitRef := t.gitRepo().
		WithNewFile("test.md", "test").
		WithExec([]string{"git", "add", "test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	actual, err := dag.GitCliff(gitRef).BumpedVersion(ctx)

	if err != nil {
		return parseErr(err)
	}

	const expected = `v1.0.1`

	if strings.TrimSpace(actual) != expected {
		return fmt.Errorf("tag does not match the expected value\nactual:   %s\nexpected: %s", actual, expected)
	}

	return err
}

// test WithBumpedVersion
func (t *Tests) WithBumpedVersion(ctx context.Context) error {

	//version should be bumped with a fix: commit
	gitRef := t.gitRepo().
		WithNewFile("test.md", "test").
		WithExec([]string{"git", "add", "test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	actual, err := dag.GitCliff(gitRef).WithBumpedVersion().Run().Stdout(ctx)

	if err != nil {
		return parseErr(err)
	}

	const expected = `v1.0.1`

	if strings.TrimSpace(actual) != expected {
		return fmt.Errorf("tag does not match the expected value\nactual:   %s\nexpected: %s", actual, expected)
	}

	return err
}

func (t *Tests) BumpedVersionIncludePath(ctx context.Context) error {

	//version should be bumped with a fix: commit.
	gitRef := t.gitRepo().
		WithNewFile("test/test.md", "test").
		WithExec([]string{"git", "add", "test/test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	actual, err := dag.GitCliff(gitRef).WithIncludePath([]string{"test/**"}).BumpedVersion(ctx)

	if err != nil {
		return parseErr(err)
	}

	const expected = `v1.0.1`

	if strings.TrimSpace(actual) != expected {
		return fmt.Errorf("tag does not match the expected value\nactual:   %s\nexpected: %s", actual, expected)
	}

	return err

}

func (t *Tests) BumpedVersionExcludePath(ctx context.Context) error {

	//fix: commit is excluded so version should NOT be bumped.
	gitRef := t.gitRepo().
		WithNewFile("test/test.md", "test").
		WithExec([]string{"git", "add", "test/test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	actual, err := dag.GitCliff(gitRef).WithExcludePath([]string{"test/**"}).BumpedVersion(ctx)

	if err != nil {
		return parseErr(err)
	}

	const expected = ``

	if strings.TrimSpace(actual) != expected {
		return fmt.Errorf("tag does not match the expected value\nactual:   %s\nexpected: %s", actual, expected)
	}

	return err
}

// test withTagPattern
func (t *Tests) WithTagPattern(ctx context.Context) error {

	//version should be bumped with a fix: commit
	gitRef := t.gitRepo().
		WithNewFile("test.md", "test").
		WithExec([]string{"git", "add", "test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	var tagPattern = []string{"v[0-9]+.[0-9]+.[0-9]+$"}

	actual, err := dag.GitCliff(gitRef).WithTagPattern(tagPattern).BumpedVersion(ctx)

	if err != nil {
		return parseErr(err)
	}

	const expected = `v1.0.1`

	if strings.TrimSpace(actual) != expected {
		return fmt.Errorf("tag does not match the expected value\nactual:   %s\nexpected: %s", actual, expected)
	}

	return err
}

// test WithLatest
func (t *Tests) WithLatest(ctx context.Context) error {

	gitRef := t.gitRepo().Directory("/repo").AsGit().Head()

	output, err := dag.GitCliff(gitRef).WithLatest().Run().Stdout(ctx)

	if err != nil {
		return parseErr(err)
	}

	if !strings.Contains(output, "## [1.0.0]") {
		return fmt.Errorf("expected output to contain '## [1.0.0]', got:\n%s", output)
	}

	return err
}

// test WithCurrent
func (t *Tests) WithCurrent(ctx context.Context) error {

	gitRef := t.gitRepo().Directory("/repo").AsGit().Head()

	output, err := dag.GitCliff(gitRef).WithCurrent().Run().Stdout(ctx)

	if err != nil {
		return parseErr(err)
	}

	if !strings.Contains(output, "## [1.0.0]") {
		return fmt.Errorf("expected output to contain '## [1.0.0]', got:\n%s", output)
	}

	return err
}

// test WithUnreleased
func (t *Tests) WithUnreleased(ctx context.Context) error {

	gitRef := t.gitRepo().Directory("/repo").AsGit().Head()

	output, err := dag.GitCliff(gitRef).WithUnreleased().Run().Stdout(ctx)

	if err != nil {
		return parseErr(err)
	}

	if !strings.Contains(output, "## [1.0.0]") {
		return fmt.Errorf("expected output to contain '## [1.0.0]', got:\n%s", output)
	}

	return err
}

// test WithOutput
func (t *Tests) WithOutput(ctx context.Context) error {

	gitRef := t.gitRepo().Directory("/repo").AsGit().Head()

	file := dag.GitCliff(gitRef).WithOutput("CHANGELOG.md").Run().File("CHANGELOG.md")

	contents, err := file.Contents(ctx)

	if err != nil {
		return parseErr(err)
	}

	if !strings.Contains(contents, "## [1.0.0]") {
		return fmt.Errorf("expected output to contain '## [1.0.0]', got:\n%s", contents)
	}

	return err
}
