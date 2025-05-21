package main

import (
	"context"
	"dagger/release/util"
	"fmt"
	"strings"
)

// Publish creates a release, uploading assets as appropriate.
func (r *Release) Publish() (string, error) {
	// use goreleaser to publish?

	return "", fmt.Errorf("not implemented")
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
	existing, err := r.existingOCITags(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("resolving existing OCI tags: %w", err)
	}

	return util.ExtraTags(version, existing)
}

// existingOCITags returns the OCI tags in a repository.
func (r *Release) existingOCITags(ctx context.Context,
	// OCI repository, e.g. localhost:5000/helloworld
	ref string,
) ([]string, error) {
	oras := dag.Container().
		From("ghcr.io/oras-project/oras:v1.2.3").
		File("/bin/oras")

	out, err := dag.Wolfi().
		Container().
		WithMountedFile("/bin/oras", oras).
		WithMountedSecret("/root/.docker/config.json", r.RegistryConfig.Secret()).
		WithExec([]string{"oras", "repo", "tags", ref}).
		Terminal().
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving repository tags: %w", err)
	}

	return strings.Fields(out), nil
}
