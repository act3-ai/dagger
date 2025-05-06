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
	"fmt"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
)

type Release struct{}

// Generate the next version from conventional commit messages (see cliff.toml). Includes 'v' prefix.
func (r *Release) Version(ctx context.Context) (string, error) {
	targetVersion, err := dag.GitCliff(r.Source).
		BumpedVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving release target version: %w", err)
	}

	return strings.TrimSpace(targetVersion), err
}

// generate Major, Minor, and latest tags based on new patch.
func (r *Release) genTagList(newVersion string, existingTags []string) ([]string, error) {

	// Filter out non-semver tags and sort
	var semverTags []string
	for _, tag := range existingTags {
		if semver.IsValid(tag) {
			semverTags = append(semverTags, tag)
		}
	}

	semver.Sort(semverTags)
	slices.Reverse(semverTags)

	// Skip tag check if newVersion is a prerelease
	if semver.Prerelease(newVersion) != "" {
		return nil, nil
	}

	// check if new tag is valid semver
	if !semver.IsValid(newVersion) {
		return nil, fmt.Errorf("new version %q is not valid semver", newVersion)
	}

	// check if new tag doesn't already exist
	for _, tag := range semverTags {
		if tag == newVersion {
			return nil, fmt.Errorf("version %s already exists", newVersion)
		}
	}

	newMajor := semver.Major(newVersion)
	newMajorMinor := semver.MajorMinor(newVersion)

	// Find latest tags for each category.
	var latestOverall, latestMajor, latestMajorMinor bool // default is false
	for _, tag := range semverTags {
		if semver.Compare(tag, newVersion) <= 0 {
			continue
		}
		if !latestOverall {
			latestOverall = true
		}
		if !latestMajor && semver.Major(tag) == newMajor {
			latestMajor = true
		}
		if !latestMajorMinor && semver.MajorMinor(tag) == newMajorMinor {
			latestMajorMinor = true
		}
		if latestOverall && latestMajor && latestMajorMinor {
			break
		}
	}

	publishTags := []string{newVersion} // Always return patch version

	if !latestMajorMinor {
		publishTags = append(publishTags, newMajorMinor)
	}
	if !latestMajor {
		publishTags = append(publishTags, newMajor)
	}
	if !latestOverall {
		publishTags = append(publishTags, "latest")
	}
	return publishTags, nil
}

// current issue with SSH AUTH SOCK: https://docs.dagger.io/api/remote-repositories/#multiple-ssh-keys-may-cause-ssh-forwarding-to-fail
func (r *Release) CheckVersion(
	ctx context.Context,
	// git repo to check existing tags
	// +default="https://github.com/dagger/dagger"
	url string,
	// ssh auth socket to use for git cloning
	sock *dagger.Socket) ([]string, error) {
	existingTags, err := dag.Git(url, dagger.GitOpts{SSHAuthSocket: sock}).Tags(ctx)
	if err != nil {
		return nil, err
	}

	newVersion, err := r.Version(ctx)
	if err != nil {
		return nil, err
	}

	return r.genTagList(newVersion, existingTags)
}
