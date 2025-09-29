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
	"dagger/act-3-pipeline/internal/dagger"
)

type Act3Pipeline struct {
	// Source git repository
	// +private
	Source *dagger.Directory
	// +private
	RegistryConfig *dagger.RegistryConfig
	// .netrc file for private modules can be passed as env var or file --netrc env:var_name, file:/filepath/.netrc
	// +optional
	// +private
	Netrc *dagger.Secret
	// +private
	GitIgnore *dagger.File
}

func New(
	// top level source code directory
	src *dagger.Directory,
	// .netrc file for private modules can be passed as env var or file --netrc env:var_name, file:/filepath/.netrc
	// +optional
	netrc *dagger.Secret,
	// Additonal .gitignore file
	// +optional
	gitIgnore *dagger.File,
) (*Act3Pipeline, error) {
	return &Act3Pipeline{
		Source:         src,
		RegistryConfig: dag.RegistryConfig(),
		Netrc:          netrc,
		GitIgnore:      gitIgnore,
	}, nil
}

type Release struct {
	// Reference to parent module instance
	pipeline *Act3Pipeline
}

// Modules for creating a release and generating associated metadata(changelog, release notes, and versioning)
func (a *Act3Pipeline) Release() *Release {
	return &Release{
		pipeline: a,
	}
}
