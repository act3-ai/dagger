package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

// This file hosts general publish functions used by project specific publish methods.

const (
	imageOras = "ghcr.io/oras-project/oras:v1.2.3"
	// A dagger module does exist for glab, but doesn't support custom base images,includes the deprecated gitlab release CLI, and doesn't have first-class support for release assets.
	// https://github.com/vbehar/daggerverse/tree/838974e23bf2afb192850103c6d76fe620f31afd/gitlab-cli
	imageGlabCLI = "registry.gitlab.com/gitlab-org/cli:latest"
)

// Publish additional tags to a remote OCI artifact.
func (r *Release) AddTags(ctx context.Context,
	// Existing OCI reference
	ref string,
	// Additional tags
	tags []string,
) (string, error) {
	return r.orasCtr().
		WithExec(append([]string{"oras", "tag", ref}, tags...)).
		Stdout(ctx)
}

// Generate extra tags based on the provided target tag.
//
// Ex: Given the patch release 'v1.2.3', with an existing 'v1.3.0' release, it returns 'v1.2'.
// Ex: Given the patch release 'v1.2.3', which is the latest and greatest, it returns 'v1', 'v1.2', 'latest'.
//
// Notice: current issue with SSH AUTH SOCK: https://docs.dagger.io/api/remote-repositories/#multiple-ssh-keys-may-cause-ssh-forwarding-to-fail
func (r *Release) ExtraTags(ctx context.Context,
	// OCI repository, e.g. localhost:5000/helloworld
	ref string,
	// target version
	version string,
) ([]string, error) {
	out, err := r.orasCtr().
		WithExec([]string{"oras", "repo", "tags", ref}).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving existing repository tags: %w", err)
	}
	existing := strings.Fields(out)

	return util.ExtraTags(version, existing)
}

// orasCtr returns a container with an oras executable, with mounted registry credentials.
func (r *Release) orasCtr() *dagger.Container {
	oras := dag.Container().
		From(imageOras).
		File("/bin/oras")

	return dag.Wolfi().
		Container().
		WithMountedFile("/bin/oras", oras).
		WithMountedSecret("/root/.docker/config.json", r.RegistryConfig.Secret())
}

// Create a release in GitHub.
func (r *Release) CreateGithub(ctx context.Context,
	// GitHub repository, without "github.com"
	repo string,
	// gitlab personal access token
	token *dagger.Secret,
	// Release version
	version string,
	// Release notes file
	notes *dagger.File,
	// Release title. Default: version
	// +optional
	title string,
	// Release assets
	// +optional
	assets []*dagger.File,
) (string, error) {

	if title == "" {
		title = version
	}

	err := dag.Gh(
		dagger.GhOpts{
			Token:  token,
			Repo:   repo,
			Source: r.gitRefAsDir(r.GitRef),
		}).
		Release().
		Create(ctx, version, title,
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
	// Release version
	version string,
	// Release notes file
	notes *dagger.File,
	// Release title. Default: version
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
		title = version
	}

	_, err = dag.Container().
		From(imageGlabCLI).
		WithMountedFile(notesFileName, notes).
		WithSecretVariable("GITLAB_TOKEN", token).
		WithEnvVariable("GITLAB_HOST", host).
		WithExec([]string{"glab", "release", "create",
			"-R", project, // repository
			version, // tag
			"--name=" + title,
			"--notes-file=" + notesFileName, // description
		}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("publishing release to '%s': %w", hostRepo, err)
	}

	if len(assets) > 0 {
		_, err := r.gitlabUploadAssets(ctx, host, project, token, version, assets)
		if err != nil {
			return "", fmt.Errorf("uploading release assets for '%s' release '%s': %w", hostRepo, version, err)
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
	// Release version
	version string,
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
					"v" + version,
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
