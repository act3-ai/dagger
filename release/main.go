// Release provides customizable release pipeline stages for Go and Python projects.
//
// Three stages are provided:
// - release {go/python} check - runs general linters and {go/python} specific linters and unit tests.
// - release prepare - generate changelog, release notes, and release version.
// - release create-{github/gitlab} - create a release page on github.com, gitlab.com, or a private GitLab instance.
//
// This module does not support image publishing, as to be flexible to
// other publishing methods; through dagger, ko, docker, etc. However, it does
// provide a couple helper functions, add-tags and extra-tags, to aid in
// publishing additional tags for an OCI image.
//
// This module does not support functional or integration testing, as such testing
// often requires extensive customization that is not easily generalized.
//
// This module uses other act3-ai modules as components, with additional functionality.
// Please refer to each modules' documentation if desired functionality is not
// available in this module.

package main

import (
	"dagger/release/internal/dagger"
)

type Release struct {
	// Git Ref Source, ex: https://gitlab.com/my/app.git
	// +private
	GitRef *dagger.GitRef
	// +private
	RegistryConfig *dagger.RegistryConfig
	// .netrc file for private modules can be passed as env var or file --netrc env:var_name, file:/filepath/.netrc
	// +optional
	// +private
	Netrc *dagger.Secret
}

func New(
	// Git Ref Source, ex: https://gitlab.com/my/app.git
	gitref *dagger.GitRef,
	// .netrc file for private modules can be passed as env var or file --netrc env:var_name, file:/filepath/.netrc
	// +optional
	netrc *dagger.Secret,
) (*Release, error) {
	return &Release{
		GitRef:         gitref,
		RegistryConfig: dag.RegistryConfig(),
		Netrc:          netrc,
	}, nil
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

// convert GitRef to a Directory
func (r *Release) gitRefAsDir(gitref *dagger.GitRef) *dagger.Directory {
	return gitref.Tree(dagger.GitRefTreeOpts{Depth: -1})
}
