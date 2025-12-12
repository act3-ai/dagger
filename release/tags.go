package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"fmt"
	"strings"

	"github.com/distribution/reference"
)

const imageOras = "ghcr.io/oras-project/oras:v1.3.0"

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

// Publish additional tags to a remote OCI artifact.
// +cache="never"
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
// +cache="never"
func (r *Release) ExtraTags(ctx context.Context,
	// OCI repository, e.g. localhost:5000/helloworld
	ref string,
	// target version
	version string,
) ([]string, error) {
	version = strings.TrimPrefix(version, "v")

	out, err := r.orasCtr().
		WithExec([]string{"oras", "repo", "tags", "--exclude-digest-tags", ref}).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving existing repository tags: %w", err)
	}
	existing := strings.Fields(out)

	return util.ExtraTags("v"+version, existing)
}

// Create extra tags based on the provided target tag.
// Combines ExtraTags() and AddTags().
// +cache="never"
func (r *Release) CreateExtraTags(ctx context.Context,
	// OCI image reference, e.g. localhost:5000/helloworld, localhost:5000/helloworld:v1.2.3
	ref string,
	// target version
	version string,
) ([]string, error) {
	named, err := reference.ParseNamed(ref)
	if err != nil {
		return nil, err
	}

	tags, err := r.ExtraTags(ctx, named.Name(), version)
	if err != nil {
		return nil, err
	}

	_, err = r.AddTags(ctx, ref, tags)
	if err != nil {
		return nil, err
	}

	fullRefs := make([]string, len(tags))
	for i, tag := range tags {
		tagged, err := reference.WithTag(named, tag)
		if err != nil {
			return nil, fmt.Errorf("failed to add tag %s to %s: %w", tag, ref, err)
		}
		fullRefs[i] = tagged.String()
	}
	return fullRefs, nil
}
