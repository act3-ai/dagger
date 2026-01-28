// A CI/Release module for releasing modules in this repo.
// This module is internal use only and not for use outside of that.

package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
	"path/filepath"
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
	repo = "act3-ai/dagger"
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
	version string,
) (string, error) {
	tag := m.Module + "/v" + version
	notesPath := m.Module + "/releases/v" + version + ".md"
	notesFile := m.GitRef.Tree().File(notesPath)

	release, err := dag.Release(m.GitRef).
		CreateGithub(ctx,
			repo,
			m.GithubToken,
			tag,
			notesFile,
			dagger.ReleaseCreateGithubOpts{Title: tag})

	if err != nil {
		return "", fmt.Errorf("%s", err)
	}

	return release, nil
}

func (m *Ci) UpgradeDagger() *dagger.Changeset {
	src := m.GitRef.Tree()
	after := dag.Container().From("registry.dagger.io/engine:latest").
		WithDirectory("/src", src).
		WithWorkdir(filepath.Join("/src", m.Module)).
		WithExec([]string{"dagger", "develop"}, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}).
		WithExec([]string{"dagger", "develop", "-m=tests"}, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}).
		Directory("/src").Filter(dagger.DirectoryFilterOpts{Gitignore: true})

	return after.Changes(src)
}
