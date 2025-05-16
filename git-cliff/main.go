// GitCliff is a highly customizable changelog generator.
package main

import (
	"context"
	"dagger/git-cliff/internal/dagger"
	"fmt"
)

const (
	imageGitCliff = "docker.io/orhunp/git-cliff" // default: "latest"
)

type GitCliff struct {
	Container *dagger.Container

	// +private
	Args []string
}

func New(
	// Git repository source.
	Source *dagger.Directory,

	// Version (image tag) to use as a git-cliff binary source.
	// +optional
	// +default="latest"
	version string,
) *GitCliff {
	return &GitCliff{
		Container: defaultContainer(Source, version),
		Args:      []string{"git-cliff"},
	}
}

// WithEnvVariable adds an environment variable to the git-cliff container.
//
// This is useful for reusability and readability by not breaking the calling chain.
func (gc *GitCliff) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,
	// The value of the environment variable (e.g., "localhost").
	value string,
	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	//
	// +optional
	expand bool,
) *GitCliff {
	gc.Container = gc.Container.WithEnvVariable(
		name,
		value,
		dagger.ContainerWithEnvVariableOpts{
			Expand: expand,
		},
	)
	return gc
}

// WithSecretVariable adds an env variable containing a secret to the git-cliff container.
//
// This is useful for reusability and readability by not breaking the calling chain.
func (gc *GitCliff) WithSecretVariable(
	// The name of the environment variable containing a secret (e.g., "PASSWORD").
	name string,
	// The value of the environment variable containing a secret.
	secret *dagger.Secret,
) *GitCliff {
	gc.Container = gc.Container.WithSecretVariable(name, secret)
	return gc
}

// Add netrc credentials.
func (gc *GitCliff) WithNetrc(
	// NETRC credentials
	netrc *dagger.Secret,
) *GitCliff {
	gc.Container = gc.Container.WithMountedSecret("/root/.netrc", netrc)
	return gc
}

// Sets the configuration file.
//
// e.g. `git-cliff --config <config>`.
func (gc *GitCliff) WithConfig(
	// git-cliff configuration file, i.e. cliff.toml.
	config *dagger.File,
) *GitCliff {
	configPath := "/work/cliff.toml"
	gc.Container = gc.Container.WithMountedFile(configPath, config)
	return gc.WithArgs("--config", configPath)
}

// Add a generic argument to pass to git-cliff.
func (gc *GitCliff) WithArg(
	// argument to pass to git-cliff
	arg string,
) *GitCliff {
	return gc.WithArgs(arg)
}

// Add generic arguments to pass to git-cliff.
func (gc *GitCliff) WithArgs(
	// arguments to pass to git-cliff
	args ...string,
) *GitCliff {
	gc.Args = append(gc.Args, args...)
	return gc
}

// Run git-cliff with all options previously provided.
//
// Run MAY be used as a "catch-all" in case functions are not implemented.
func (gc *GitCliff) Run(
	// arguments and flags, without `git-cliff`
	// +optional
	args []string,
) *dagger.Container {
	gc = gc.WithArgs(args...)
	return gc.Container.WithExec(gc.Args)
}

// Prints bumped version for unreleased changes.
func (gc *GitCliff) BumpedVersion(ctx context.Context) (string, error) {
	return gc.
		WithArgs("--bumped-version").
		Container.
		WithExec(gc.Args).
		Stdout(ctx)
}

// Sets the GitHub API token.
//
// e.g. `GITHUB_TOKEN=<token> git-cliff`.
func (gc *GitCliff) WithGithubToken(
	// GitHub API token.
	token *dagger.Secret,
) *GitCliff {
	return gc.WithSecretVariable("GITHUB_TOKEN", token)
}

// Sets the GitLab API token.
//
// e.g. `GITLAB_TOKEN=<token> git-cliff`.
func (gc *GitCliff) WithGitlabToken(
	// GitLab API token.
	token *dagger.Secret,
) *GitCliff {
	return gc.WithSecretVariable("GITLAB_TOKEN", token)
}

// Sets the Gitea API token.
//
// e.g. `GITEA_TOKEN=<token> git-cliff`.
func (gc *GitCliff) WithGiteaToken(
	// Gitea API token.
	token *dagger.Secret,
) *GitCliff {
	return gc.WithSecretVariable("GITEA_TOKEN", token)
}

// Bump the version for unreleased changes. Optionally with specified version.
//
// e.g. `git-cliff --bump`.
func (gc *GitCliff) WithBump(
	// specified version bump. git-cliff defaults to "auto"
	// possible values: auto, major, minor, patch.
	// +optional
	version string,
) *GitCliff {
	gc = gc.WithArgs("--bump")
	if version != "" {
		gc = gc.WithArgs(version)
	}
	return gc
}

// Processes the commits starting from the latest tag.
//
// e.g. `git-cliff --latest`.
func (gc *GitCliff) WithLatest() *GitCliff {
	return gc.WithArgs("--latest")
}

// Processes the commits that belog to the current tag.
//
// e.g. `git-cliff --current`
func (gc *GitCliff) WithCurrent() *GitCliff {
	return gc.WithArgs("--current")
}

// Processes the commits that do not belog to a tag.
//
// e.g. `git-cliff --unreleased`.
func (gc *GitCliff) WithUnreleased() *GitCliff {
	return gc.WithArgs("--unreleased")
}

// Sorts the tags topologically.
//
// e.g. `git-cliff --topo-order`.
func (gc *GitCliff) WithTopoOrder() *GitCliff {
	return gc.WithArgs("--topo-order")
}

// Sets the git repository.
//
// e.g. `git-cliff --repository <repo>...`.
func (gc *GitCliff) WithRepository(
	// git repository (one or more)
	repo []string,
) *GitCliff {
	return gc.
		WithArgs("--repository").
		WithArgs(repo...)
}

// Sets comits that will be skipped in the changelog.
//
// e.g. `git-cliff --skip-commit <sha1>...`.
func (gc *GitCliff) WithSkipCommit(
	// Commits
	sha1 []string,
) *GitCliff {
	return gc.
		WithArgs("--skip-commit").
		WithArgs(sha1...)
}

// Prepends entries to the given changelog file.
//
// e.g. `git-cliff --prepend <changelog>`.
func (gc *GitCliff) WithPrepend(
	// Path to changelog, relative to source git directory
	changelog string,
) *GitCliff {
	return gc.WithArgs("--prepend", changelog)
}

// Writes output to the fiven file.
//
// e.g. `git-cliff --output <path>`.
func (gc *GitCliff) WithOutput(
	// Write output to file, relative to source git directory.
	path string,
) *GitCliff {
	return gc.WithArgs("--output", path)
}

// Strips the given parts from the changelog.
//
// e.g. `git-cliff --strip <part>`.
func (gc *GitCliff) WithStrip(
	// Part of changelog to strip. Possible values: header, footer, all.
	part string,
) *GitCliff {
	return gc.WithArgs("--strip", part)
}

// defaultContainer constructs a minimal container containing a source git repository.
func defaultContainer(source *dagger.Directory, version string) *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("%s:%s", imageGitCliff, version)).
		WithWorkdir("/work/src").
		WithMountedDirectory("/work/src", source)
}
