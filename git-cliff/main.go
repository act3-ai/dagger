// GitCliff is a highly customizable changelog generator.
package main

import (
	"context"
	"dagger/git-cliff/internal/dagger"
	"fmt"
	"strings"
)

const (
	imageGitCliff = "docker.io/orhunp/git-cliff" // default: "latest"
)

type GitCliff struct {
	// +private
	Gitref *dagger.GitRef
	// +private
	Options *Options
}

// Options represents all configurable options for running git-cliff
type Options struct {
	GitCliffVersion string
	Netrc           *dagger.Secret
	Tag             string
	TagPattern      []string
	Bump            bool
	BumpedVersion   bool
	Unreleased      bool
	Strip           string
	OutputFile      string
	Prepend         string
	GithubToken     *dagger.Secret
	GitlabToken     *dagger.Secret
	GiteaToken      *dagger.Secret
	Config          string
	IncludePath     []string
	ExcludePath     []string
	SkipCommits     []string
	Latest          bool
	Current         bool
	TopoOrder       bool
}

func New(ctx context.Context,
	// Git repository source.
	gitref *dagger.GitRef,
	// Version of git-cliff image.
	// +optional
	// +default=latest
	gitCliffVersion string,
	// Mount netrc credentials.
	// +optional
	netrc *dagger.Secret,
	// Sets the tag for the latest version
	// +optional
	tag string,
	// Sets the regex for matching git tags
	// +optional
	tagPattern []string,
	// Bumps the version for unreleased changes. Optionally with specified version
	// +optional
	bump bool,
	// Prints bumped version for unreleased changes
	// +optional
	bumpedVersion bool,
	//processes the commits that do not belong to a tag
	// +optional
	// +default=true
	unreleased bool,
	// Strips the given parts from the changelog [possible values: header, footer, all]
	// +optional
	strip string,
	//writes output to the given file
	// +optional
	outputFile string,
	//Prepends entries to the given changelog file
	// +optional
	prepend string,
	// private token to use when authenticating with a private github instance set in cliff.toml
	// See: https://git-cliff.org/docs/integration/github
	// +optional
	githubToken *dagger.Secret,
	// private token to use when authenticating with a private gitlab instance set in cliff.toml
	// See: https://git-cliff.org/docs/integration/gitlab
	// +optional
	gitlabToken *dagger.Secret,
	// private token to use when authenticating with a private gitea instance set in cliff.toml
	// See: https://git-cliff.org/docs/integration/gitea
	// +optional
	giteaToken *dagger.Secret,
	// path to a git-cliff config. Defaults to root directory of given git-ref source
	// +optional
	config string,
	// Sets the path to include related commits
	// +optional
	includePath []string,
	// Sets the path to exnclude related commits
	// +optional
	excludePath []string,
	// Sets commits that will be skipped in the changelog
	// +optional
	skipCommits []string,
	// Processes the commits starting from the latest tag
	// +optional
	latest bool,
	// Processes the commits that belong to the current tag
	// +optional
	current bool,
	// Sorts the tags topologically
	// +optional
	topoOrder bool,
) *GitCliff {
	return &GitCliff{
		Gitref: gitref,
		Options: &Options{
			GitCliffVersion: gitCliffVersion,
			Netrc:           netrc,
			Tag:             tag,
			TagPattern:      tagPattern,
			Bump:            bump,
			BumpedVersion:   bumpedVersion,
			Unreleased:      unreleased,
			Strip:           strip,
			OutputFile:      outputFile,
			Prepend:         prepend,
			GithubToken:     githubToken,
			GitlabToken:     gitlabToken,
			GiteaToken:      giteaToken,
			Config:          config,
			IncludePath:     includePath,
			ExcludePath:     excludePath,
			SkipCommits:     skipCommits,
			Latest:          latest,
			Current:         current,
			TopoOrder:       topoOrder,
		},
	}
}

// Run git-cliff with all options provided.
//
// Run MAY be used as a "catch-all" in case functions are not implemented.
func (gc *GitCliff) Run() *dagger.Container {
	//convert git-ref to a *dagger.Directory
	gitRefDir := gc.Gitref.Tree(dagger.GitRefTreeOpts{Depth: -1})
	//default git-cliff cmd
	cmd := []string{"git-cliff", "--use-native-tls"}
	srcDir := "/work/src"
	//base git-cliff container
	ctr := dag.Container().
		From(fmt.Sprintf("%s:%s", imageGitCliff, gc.Options.GitCliffVersion)).
		WithMountedDirectory(srcDir, gitRefDir).
		WithWorkdir(srcDir)

	// parse given options set
	if gc.Options.Netrc != nil {
		ctr = ctr.WithMountedSecret("/root/.netrc", gc.Options.Netrc)
	}

	if gc.Options.Tag != "" {
		cmd = append(cmd, "--tag", gc.Options.Tag)
	}

	if len(gc.Options.TagPattern) > 0 {
		cmd = append(cmd, "--tag-pattern", strings.Join(gc.Options.TagPattern, ","))
	}

	if gc.Options.Bump {
		cmd = append(cmd, "--bump")
	}

	if gc.Options.BumpedVersion {
		cmd = append(cmd, "--bumped-version")
	}

	if gc.Options.Unreleased {
		cmd = append(cmd, "--unreleased")
	}

	if gc.Options.Strip != "" {
		cmd = append(cmd, "--strip", gc.Options.Strip)
	}

	if gc.Options.OutputFile != "" {
		cmd = append(cmd, "--output", gc.Options.OutputFile)
	}

	if gc.Options.Prepend != "" {
		cmd = append(cmd, "--prepend", gc.Options.Prepend)
	}

	if gc.Options.GithubToken != nil {
		ctr = ctr.
			WithSecretVariable("GITHUB_TOKEN", gc.Options.GithubToken)
	}
	if gc.Options.GitlabToken != nil {
		ctr = ctr.
			WithSecretVariable("GITLAB_TOKEN", gc.Options.GitlabToken)
	}
	if gc.Options.GiteaToken != nil {
		ctr = ctr.
			WithSecretVariable("GITEA_TOKEN", gc.Options.GiteaToken)
	}

	if gc.Options.Config != "" {
		cmd = append(cmd, "--config", gc.Options.Config)
	}

	if len(gc.Options.IncludePath) > 0 {
		cmd = append(cmd, "--include-path", strings.Join(gc.Options.IncludePath, ","))
	}

	if len(gc.Options.ExcludePath) > 0 {
		cmd = append(cmd, "--exclude-path", strings.Join(gc.Options.ExcludePath, ","))
	}

	if gc.Options.Latest {
		cmd = append(cmd, "--latest")
	}

	if gc.Options.Current {
		cmd = append(cmd, "--current")
	}

	if gc.Options.TopoOrder {
		cmd = append(cmd, "--topo-order")
	}

	// cmd := gc.Command
	// cmd = append(cmd, args...)
	// return gc.Container.WithExec(cmd)

	return ctr.WithExec(cmd)
}

// Prints bumped version for unreleased changes.
func (gc *GitCliff) BumpedVersion(ctx context.Context,
	// Configuration file path in provided git repository/ref.
	// +optional
	config string,
	// additional arguments and flags for git-cliff
	// +optional
	args []string,
) (string, error) {
	cmd := gc.Command
	cmd = append(cmd, "--bumped-version")
	cmd = append(cmd, args...)

	if config != "" {
		cmd = append(cmd, "--config", config)
	}

	ctr := gc.Container.WithExec(cmd)
	// The check below is needed due to how git-cliff returns its warning/error logs.
	//  Warnings are returned as errors and not stdout
	stderr, err := ctr.Stderr(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting version: %w", err)
	}

	if stderr != "" {
		// git-cliff returns the latest tag it found when there is nothing to bump
		// This will return an empty string instead in that case
		if strings.Contains(stderr, "There is nothing to bump") {
			return "", nil
		}

		if strings.Contains(stderr, "ERROR") {
			return "", fmt.Errorf("error getting version: %s", stderr)
		}

	}

	stdout, err := ctr.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting version: %w", err)
	}

	return stdout, nil
}
