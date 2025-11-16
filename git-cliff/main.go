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
	//Returns a git-cliff container with optionally mounted secret variables
	//for private gitlab, github, or gitea tokens
	Container *dagger.Container
	// +private
	Gitref *dagger.GitRef
	// +private
	Command []string
}

func New(ctx context.Context,
	// Git repository source.
	gitRef *dagger.GitRef,
	// Version of git-cliff image.
	// +optional
	// +default=latest
	gitCliffVersion string,
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
) *GitCliff {

	//convert git-ref to a *dagger.Directory
	gitRefDir := gitRef.Tree(dagger.GitRefTreeOpts{Depth: -1})
	//default git-cliff cmd
	cmd := []string{"git-cliff", "--use-native-tls"}
	srcDir := "/work/src"
	//base git-cliff container
	ctr := dag.Container().
		From(fmt.Sprintf("%s:%s", imageGitCliff, gitCliffVersion)).
		WithMountedDirectory(srcDir, gitRefDir).
		WithWorkdir(srcDir)

	if githubToken != nil {
		ctr = ctr.
			WithSecretVariable("GITHUB_TOKEN", githubToken)
	}

	if gitlabToken != nil {
		ctr = ctr.
			WithSecretVariable("GITLAB_TOKEN", gitlabToken)
	}
	if giteaToken != nil {
		ctr = ctr.
			WithSecretVariable("GITEA_TOKEN", giteaToken)
	}

	return &GitCliff{
		Container: ctr,
		Command:   cmd,
		Gitref:    gitRef,
	}
}

// generate a changelog file with unreleased changes and bumps the tag if tag is not provided
// If file already exists, it will prepend to the existing changelog instead of creating a new one.
func (gc *GitCliff) Changelog(
	ctx context.Context,
	//file path to output or prepend generated changelog.
	// +optional
	// +default="CHANGELOG.md"
	changelog string,
	//tag to generate changelog for
	// +optional
	tag string,
	//path to git-cliff config to use for generating changelog in provided git-ref source.
	// +optional
	// +default="cliff.toml"
	config string,
) *dagger.File {
	cmd := gc.Command
	cmd = append(cmd,
		"--config", config,
		"--unreleased",
		"--strip",
		"footer",
	)

	//use provided tag, otherwise bump automatically
	if tag != "" {
		cmd = append(cmd, "--tag", tag)
	} else {
		cmd = append(cmd, "--bump")
	}
	//check if changelog exists and either prepend or generate new changelog file
	exists, err := gc.Gitref.Tree().Exists(ctx, changelog)
	if err != nil {
		panic(fmt.Errorf("failed to check if %s exists: %w", changelog, err))
	}

	if exists {
		cmd = append(cmd, "--prepend", changelog)
	} else {
		cmd = append(cmd, "--output", changelog)
	}

	return gc.Container.WithExec(cmd).File(changelog)
}

// generate release notes file with unreleased changes and bumps the tag if tag is not provided
func (gc *GitCliff) ReleaseNotes(
	ctx context.Context,
	//file path to output release notes.
	// +optional
	// +default="releasenotes.md"
	notes string,
	//tag to generate changelog for
	// +optional
	tag string,
	//path to git-cliff config to use for generating changelog in provided git-ref source.
	// +optional
	// +default="cliff.toml"
	config string,
	// append additional provided release notes
	// +optional
	extraNotes string,
) *dagger.File {
	cmd := gc.Command
	cmd = append(cmd,
		"--config", config,
		"--unreleased",
		"--strip",
		"all",
	)

	//use provided tag, otherwise bump automatically
	if tag != "" {
		cmd = append(cmd, "--tag", tag)
	} else {
		cmd = append(cmd, "--bump")
	}

	//generate release notes and append any extraNotes provided
	releaseNotes, err := gc.Container.WithExec(cmd).Stdout(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to generate release notes: %w", err))
	}

	if extraNotes != "" {
		releaseNotes = strings.Replace(releaseNotes, "###", extraNotes+"\n###", 1)
	}

	return gc.Container.WithExec(cmd).WithNewFile(notes, releaseNotes).File(notes)
}

// Prints a bumped tag for unreleased changes.
func (gc *GitCliff) BumpedVersion(ctx context.Context,
	//path to git-cliff config to use for generating changelog in provided git-ref source.
	// +optional
	// +default="cliff.toml"
	config string,
) (string, error) {
	cmd := gc.Command
	cmd = append(cmd,
		"--config", config,
		"--unreleased",
		"--bumped-version",
	)

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
