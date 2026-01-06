// A CI/Release module for releasing modules in this repo.
// This module is internal use only and not for use outside of that.

package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
)

type Ci struct {
	GitRef      *dagger.GitRef
	GithubToken *dagger.Secret
	Module      string
}

func New(
	// +defaultPath="."
	gitRef *dagger.GitRef,
	githubToken *dagger.Secret,
	module string,
) *Ci {

	return &Ci{
		GitRef:      gitRef,
		GithubToken: githubToken,
		Module:      module,
	}
}

const (
	repo      = "act3-ai/dagger"
	notesPath = "/images.tar"
)

func (m *Ci) Prepare(ctx context.Context,
) (*dagger.Changeset, error) {

	version, err := dag.Release(m.GitRef).
		Version(ctx, dagger.ReleaseVersionOpts{WorkingDir: m.Module, GithubToken: m.GithubToken})
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	fmt.Printf("Version bump found: %s", version)

	release := dag.Release(m.GitRef).
		Prepare(version, dagger.ReleasePrepareOpts{WorkingDir: m.Module, GithubToken: m.GithubToken})
	return release, nil
}

func (m *Ci) Release(ctx context.Context,
	tag string,
	notes *dagger.File,
	title string,
) (string, error) {

	release, err := dag.Release(m.GitRef).
		CreateGithub(ctx, repo, m.GithubToken, tag, notes, dagger.ReleaseCreateGithubOpts{Title: title})
	if err != nil {
		return "", fmt.Errorf("%s", err)
	}
	return release, nil
}

func (m *Ci) UpgradeDagger(ctx context.Context,

) (string, error) {
	daggerVersion, err := dag.Container().From("registry.dagger.io/engine:latest").
		WithDirectory("/src", m.GitRef.Tree(dagger.GitRefTreeOpts{Depth: -1})).Terminal().
		WithExec([]string{"sh", "-c",
			"dagger --silent version | cut -f 2 -d ' '"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	return daggerVersion, nil
}
