// A generated module for Release functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
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
	// top level source code git directory
	src *dagger.Directory,
) *Release {
	return &Release{
		Source:         src,
		RegistryConfig: dag.RegistryConfig(),
	}
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
	// git repo to check existing tags
	url string,
	// ssh auth socket to use for git cloning
	sock *dagger.Socket,
	// target version
	// +optional
	version string,
) ([]string, error) {
	existingTags, err := dag.Git(url, dagger.GitOpts{SSHAuthSocket: sock}).Tags(ctx)
	if err != nil {
		return nil, err
	}

	newVersion, err := r.Version(ctx)
	if err != nil {
		return nil, err
	}

	return util.ExtraTags(newVersion, existingTags)
}
