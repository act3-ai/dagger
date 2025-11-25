package main

import (
	"context"
	"dagger/release/internal/dagger"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

// This file hosts general publish functions used by project specific publish methods.

// A dagger module does exist for glab, but doesn't support custom base images,includes the deprecated gitlab release CLI, and doesn't have first-class support for release assets.
// https://github.com/vbehar/daggerverse/tree/838974e23bf2afb192850103c6d76fe620f31afd/gitlab-cli
const imageGlabCLI = "registry.gitlab.com/gitlab-org/cli:latest"

// Create a release in GitHub.
func (r *Release) CreateGithub(ctx context.Context,
	// GitHub repository, without "github.com"
	repo string,
	// gitlab personal access token
	token *dagger.Secret,
	// tag to create release with
	tag string,
	// Release notes file
	notes *dagger.File,
	// Release title. Default: tag
	// +optional
	title string,
	// Release assets
	// +optional
	assets []*dagger.File,
) (string, error) {

	err := dag.Gh(
		dagger.GhOpts{
			Token:  token,
			Repo:   repo,
			Source: r.GitRef.Tree(),
		}).
		Release().
		Create(ctx, tag, title,
			dagger.GhReleaseCreateOpts{
				NotesFile: notes,
				Files:     assets,
			})
	if err != nil {
		return "", fmt.Errorf("publishing release to 'github.com/%s': %w", repo, err)
	}
	return fmt.Sprintf("Successfully published release to 'github.com/%s'", repo), nil
}

// Create a release in a public or private GitLab instance.
func (r *Release) CreateGitlab(ctx context.Context,
	// GitLab host
	// +optional
	// +default="gitlab.com"
	host string,
	// GitLab repository, without host.
	project string,
	// GitLab personal access token
	token *dagger.Secret,
	// Release tag
	tag string,
	// Release notes file
	notes *dagger.File,
	// Release title. Default: tag
	// +optional
	title string,
	// Release assets
	// +optional
	assets []*dagger.File,
) (string, error) {

	notesFileName, err := notes.Name(ctx)
	if err != nil {
		return "", err
	}

	hostRepo := fmt.Sprintf("%s/%s", host, project)
	if title == "" {
		title = tag
	}

	_, err = dag.Container().
		From(imageGlabCLI).
		WithMountedFile(notesFileName, notes).
		WithSecretVariable("GITLAB_TOKEN", token).
		WithEnvVariable("GITLAB_HOST", host).
		WithExec([]string{"glab", "release", "create",
			"-R=" + project, // repository
			tag,             // tag
			"--name=" + title,
			"--notes-file=" + notesFileName, // description
		}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("publishing release to '%s': %w", hostRepo, err)
	}

	if len(assets) > 0 {
		_, err := r.gitlabUploadAssets(ctx, host, project, token, tag, assets)
		if err != nil {
			return "", fmt.Errorf("uploading release assets for '%s' release '%s': %w", hostRepo, tag, err)
		}
	}

	return fmt.Sprintf("Successfully published release to '%s'", hostRepo), nil
}

// gitlabUploadAssets uploads assets to an existing release tag on GitLab.
func (r *Release) gitlabUploadAssets(ctx context.Context,
	// GitLab host
	host string,
	// GitLab project (repository)
	project string,
	// GitLab personal access token
	token *dagger.Secret,
	// Release tag
	tag string,
	// Release assets
	assets []*dagger.File,
) (string, error) {
	ctx, span := Tracer().Start(ctx, "Upload Assets")
	defer span.End()

	p := pool.NewWithResults[string]().WithContext(ctx)
	for _, asset := range assets {
		p.Go(func(ctx context.Context) (string, error) {
			assetName, err := asset.Name(ctx)
			if err != nil {
				return "", err
			}
			_, err = dag.Container().
				From(imageGlabCLI).
				WithMountedFile(assetName, asset).
				WithSecretVariable("GITLAB_TOKEN", token).
				WithEnvVariable("GITLAB_HOST", host).
				WithExec([]string{"glab", "release", "upload",
					"-R", project,
					tag,
					assetName},
				).
				Stdout(ctx)

			if err != nil {
				return "", fmt.Errorf("uploading release asset %s: %w", assetName, err)
			}
			return fmt.Sprintf("Asset Uploaded - %s", assetName), nil
		})
	}

	result, err := p.Wait()
	return strings.Join(result, "\n"), err
}
