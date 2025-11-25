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
		WithFile("cliff.toml", dag.CurrentModule().Source().File("test.cliff.toml")).
		WithExec([]string{"git", "add", "README.md", "cliff.toml"}).
		WithExec([]string{"git", "commit", "-m", "fix: Initial commit"}).
		WithExec([]string{"git", "tag", "-a", "-m", "Initial commit", "v1.0.0"})

}

// Run all tests
func (t *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(t.BumpedVersion)
	p.Go(t.ChangelogOutput)
	p.Go(t.ChangelogPrepend)
	p.Go(t.ReleaseNotes)
	p.Go(t.ReleaseNotesWithExtraNotes)

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

// test changelog output
func (t *Tests) ChangelogOutput(ctx context.Context) error {

	gitRef := t.gitRepo().
		WithNewFile("test.md", "git-cliff test").
		WithExec([]string{"git", "add", "test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	const expected = `## [1.0.1]

### 🐛 Bug Fixes

- Test tag
`

	actual, err := dag.GitCliff(gitRef).Changelog().Contents(ctx)

	if err != nil {
		return parseErr(err)
	}
	if expected != actual {
		return fmt.Errorf("unexpected patch\nACTUAL:\n%s\nEXPECTED:\n%s\n", actual, expected)
	}

	return err
}

// test changelog prepend
func (t *Tests) ChangelogPrepend(ctx context.Context) error {

	gitRef := t.gitRepo().
		WithNewFile("CHANGELOG.md", "## [1.0.0]").
		WithExec([]string{"git", "add", "CHANGELOG.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	const expected = `## [1.0.1]

### 🐛 Bug Fixes

- Test tag
## [1.0.0]`

	actual, err := dag.GitCliff(gitRef).Changelog().Contents(ctx)

	if err != nil {
		return parseErr(err)
	}
	if expected != actual {
		return fmt.Errorf("unexpected patch\nACTUAL:\n%s\nEXPECTED:\n%s\n", actual, expected)
	}

	return err
}

// test releasenotes
func (t *Tests) ReleaseNotes(ctx context.Context) error {

	gitRef := t.gitRepo().
		WithNewFile("test.md", "git-cliff test").
		WithExec([]string{"git", "add", "test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	const expected = `## [1.0.1]

### 🐛 Bug Fixes

- Test tag
`

	actual, err := dag.GitCliff(gitRef).ReleaseNotes().Contents(ctx)

	if err != nil {
		return parseErr(err)
	}
	if expected != actual {
		return fmt.Errorf("unexpected patch\nACTUAL:\n%s\nEXPECTED:\n%s\n", actual, expected)
	}

	return err
}

// test releasenotes with extra notes added
func (t *Tests) ReleaseNotesWithExtraNotes(ctx context.Context) error {

	gitRef := t.gitRepo().
		WithNewFile("test.md", "git-cliff test").
		WithExec([]string{"git", "add", "test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	const expected = `## [1.0.1]

extra notes
### 🐛 Bug Fixes

- Test tag
`

	actual, err := dag.GitCliff(gitRef).ReleaseNotes(dagger.GitCliffReleaseNotesOpts{ExtraNotes: "extra notes"}).Contents(ctx)

	if err != nil {
		return parseErr(err)
	}
	if expected != actual {
		return fmt.Errorf("unexpected patch\nACTUAL:\n%s\nEXPECTED:\n%s\n", actual, expected)
	}

	return err
}
