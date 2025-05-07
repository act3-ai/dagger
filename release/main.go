// A generated module for Release functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return rypes using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"fmt"
	"strings"

	"oras.land/oras-go/v2/registry/remote"
)

type Release struct {
	// Source git repository
	Source *dagger.Directory

	// +private
	RegistryConfig *dagger.RegistryConfig
	// +private
	Netrc *dagger.Secret
}

func New(
	// top level source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) *Release {
	return &Release{
		Source:         src,
		RegistryConfig: dag.RegistryConfig(),
	}
}

// Add credentials for a private registry.
func (r *Release) WithRegistryAuth(
	// registry's hostname
	address string,
	// username in registry
	username string,
	// password or token for registry
	secret *dagger.Secret,
) *Release {
	r.RegistryConfig = r.RegistryConfig.WithRegistryAuth(address, username, secret)
	return r
}

// Removes credentials for a private registry.
func (r *Release) WithoutRegistryAuth(
	// registry's hostname
	address string,
) *Release {
	r.RegistryConfig = r.RegistryConfig.WithoutRegistryAuth(address)
	return r
}

// Add netrc credentials for a private git repository.
func (r *Release) WithNetrc(
	// NETRC credentials
	netrc *dagger.Secret,
) *Release {
	r.Netrc = netrc
	return r
}

// Generate the next version from conventional commit messages (see cliff.toml).
//
// Includes 'v' prefix.
func (r *Release) Version(ctx context.Context) (string, error) {
	targetVersion, err := dag.GitCliff(r.Source).
		BumpedVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving release target version: %w", err)
	}

	return strings.TrimSpace(targetVersion), err
}

// Generate extra tags based on the provided target tag.
//
// Ex: Given the patch release 'v1.2.3', with an existing 'v1.3.0' release, it returns 'v1.2'.
// Ex: Given the patch release 'v1.2.3', which is the latest and greatest, it returns 'v1', 'v1.2', 'latest'.
//
// Notice: current issue with SSH AUTH SOCK: https://docs.dagger.io/api/remote-repositories/#multiple-ssh-keys-may-cause-ssh-forwarding-to-fail
func (r *Release) ExtraTags(
	ctx context.Context,
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
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("initializing repository: %w", err)
	}

	var ociTags []string
	err = repo.Tags(ctx, "", func(tags []string) error {
		ociTags = append(ociTags, tags...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fetching repository tags: %w", err)
	}

	return ociTags, nil
}
