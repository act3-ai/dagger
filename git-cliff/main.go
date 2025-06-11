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
	Flags []string
}

func New(ctx context.Context,
	// Git repository source.
	src *dagger.Directory,

	// Custom container to use as a base container. Must have 'git-cliff' available on PATH.
	// +optional
	container *dagger.Container,

	// Version (image tag) to use as a git-cliff binary source.
	// +optional
	// +default="latest"
	version string,

	// Configuration file.
	// +optional
	config *dagger.File,

	// Mount netrc credentials for a private git repository.
	// +optional
	netrc *dagger.Secret,
) *GitCliff {
	if container == nil {
		container = defaultContainer(version)
	}

	flags := []string{"git-cliff", "--use-native-tls"}
	srcDir := "/work/src"
	container = container.With(
		func(c *dagger.Container) *dagger.Container {
			if config != nil {
				cfgPath, err := config.Name(ctx)
				if err != nil {
					panic(fmt.Errorf("resolving configuration file name: %w", err))
				}
				c = c.WithMountedFile(cfgPath, config)
				flags = append(flags, "--config", cfgPath)
			}
			return c
		}).With(
		func(c *dagger.Container) *dagger.Container {
			if netrc != nil {
				c = c.WithMountedSecret("/root/.netrc", netrc)
			}
			return c
		}).
		WithWorkdir(srcDir).
		WithMountedDirectory(srcDir, src)

	return &GitCliff{
		Container: container,
		Flags:     flags,
	}
}

// WithPrivateGitlabHost provides conveneince for using git-cliff with a private GitLab host. Altenatively, use WithEnvVariable and WithSecretVariable as needed.
//
// Sets GITLAB_API_URL, GITLAB_REPO, and GITLAB_TOKEN.
func (gc *GitCliff) WithPrivateGitlabHost(
	// API URL, typically https://<host>/api/v4
	apiURL string,
	// Repository
	repo string,
	// Access token
	token *dagger.Secret,
) *GitCliff {
	gc.Container = gc.Container.WithEnvVariable("GITLAB_API_URL", apiURL).
		WithEnvVariable("GITLAB_REPO", repo).
		WithSecretVariable("GITLAB_TOKEN", token)

	return gc
}

// WithPrivateGiteaHost provides conveneince for using git-cliff with a private Gitea host.Altenatively, use WithEnvVariable and WithSecretVariable as needed.
//
// Sets GITEA_API_URL, GITEA_REPO, and GITEA_TOKEN.
func (gc *GitCliff) WithPrivateGiteaHost(
	// API URL, typically https://<host>/api/v4
	apiURL string,
	// Repository
	repo string,
	// Access token
	token *dagger.Secret,
) *GitCliff {
	gc.Container = gc.Container.WithEnvVariable("GITEA_API_URL", apiURL).
		WithEnvVariable("GITEA_REPO", repo).
		WithSecretVariable("GITEA_TOKEN", token)

	return gc
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

// Run git-cliff with all options previously provided.
//
// Run MAY be used as a "catch-all" in case functions are not implemented.
func (gc *GitCliff) Run(
	// arguments and flags, without `git-cliff`
	// +optional
	args []string,
) *dagger.Container {
	gc.Flags = append(gc.Flags, args...)
	return gc.Container.WithExec(gc.Flags)
}

// Prints bumped version for unreleased changes.
func (gc *GitCliff) BumpedVersion(ctx context.Context,
	// additional arguments and flags for git-cliff
	// +optional
	args []string) (string, error) {
	gc.Flags = append(gc.Flags, "--bumped-version")

	gc.Flags = append(gc.Flags, args...)
	return gc.Container.WithExec(gc.Flags).
		Stdout(ctx)
}

// Generate a changelog for a specific version, ignoring configured method of version bumping.
//
// e.g. `git-cliff --tag <version>`.
func (gc *GitCliff) WithTag(ctx context.Context,
	// Specific tag
	version string,
) *GitCliff {
	gc.Flags = append(gc.Flags, "--tag", version)
	return gc
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

// Bump the version for unreleased changes. Optionally with specified bump method/type.
//
// e.g. `git-cliff --bump`.
func (gc *GitCliff) WithBump(
	// bump method/type. Supported values: 'major', 'minor', and 'patch'.
	// +optional
	method string,
) *GitCliff {
	gc.Flags = append(gc.Flags, "--bump")
	if method != "" {
		gc.Flags = append(gc.Flags, method)
	}
	return gc
}

// Processes the commits starting from the latest tag.
//
// e.g. `git-cliff --latest`.
func (gc *GitCliff) WithLatest() *GitCliff {
	gc.Flags = append(gc.Flags, "--latest")
	return gc
}

// Processes the commits that belog to the current tag.
//
// e.g. `git-cliff --current`
func (gc *GitCliff) WithCurrent() *GitCliff {
	gc.Flags = append(gc.Flags, "--current")
	return gc
}

// Processes the commits that do not belog to a tag.
//
// e.g. `git-cliff --unreleased`.
func (gc *GitCliff) WithUnreleased() *GitCliff {
	gc.Flags = append(gc.Flags, "--unreleased")
	return gc
}

// Sorts the tags topologically.
//
// e.g. `git-cliff --topo-order`.
func (gc *GitCliff) WithTopoOrder() *GitCliff {
	gc.Flags = append(gc.Flags, "--topo-order")
	return gc
}

// Sets the git repository.
//
// e.g. `git-cliff --repository <repo>...`.
func (gc *GitCliff) WithRepository(
	// git repository (one or more)
	repo []string,
) *GitCliff {
	for _, r := range repo {
		gc.Flags = append(gc.Flags, "--repository", r)
	}
	return gc
}

// Sets the path to include related commits.
//
// e.g. `git-cliff --include-pattern <pattern>...`.
func (gc *GitCliff) WithIncludePath(
	// glob pattern or direct path (one or more)
	pattern []string,
) *GitCliff {
	for _, p := range pattern {
		gc.Flags = append(gc.Flags, "--include-path", p)
	}
	return gc
}

// Sets the path to exclude related commits.
//
// e.g. `git-cliff --include-pattern <pattern>...`.
func (gc *GitCliff) WithExcludePath(
	// glob pattern or direct path (one or more)
	pattern []string,
) *GitCliff {
	for _, p := range pattern {
		gc.Flags = append(gc.Flags, "--exclude-path", p)
	}
	return gc
}

// Sets comits that will be skipped in the changelog.
//
// e.g. `git-cliff --skip-commit <sha1>...`.
func (gc *GitCliff) WithSkipCommit(
	// Commits (one or more)
	sha1 []string,
) *GitCliff {
	for _, commit := range sha1 {
		gc.Flags = append(gc.Flags, "--skip-commit", commit)
	}
	return gc
}

// Prepends entries to the given changelog file.
//
// e.g. `git-cliff --prepend <changelog>`.
func (gc *GitCliff) WithPrepend(
	// Path to changelog, relative to source git directory
	changelog string,
) *GitCliff {
	gc.Flags = append(gc.Flags, "--prepend", changelog)
	return gc
}

// Writes output to the fiven file.
//
// e.g. `git-cliff --output <path>`.
func (gc *GitCliff) WithOutput(
	// Write output to file, relative to source git directory.
	path string,
) *GitCliff {
	gc.Flags = append(gc.Flags, "--output", path)
	return gc
}

// Strips the given parts from the changelog.
//
// e.g. `git-cliff --strip <part>`.
func (gc *GitCliff) WithStrip(
	// Part of changelog to strip. Supported values: header, footer, all.
	part string,
) *GitCliff {
	gc.Flags = append(gc.Flags, "--strip", part)
	return gc
}

// defaultContainer constructs a minimal container containing a source git repository.
func defaultContainer(version string) *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("%s:%s", imageGitCliff, version))
}
