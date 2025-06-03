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
	"dagger/release/internal/dagger"
)

type Release struct {
	// Source git repository
	// +private
	Source *dagger.Directory
	// +private
	RegistryConfig *dagger.RegistryConfig
	// .netrc file for private modules can be passed as env var or file --netrc env:var_name, file:/filepath/.netrc
	// +optional
	// +private
	Netrc *dagger.Secret
	// TODO: add optional overrides for disabling default behavior
	// +private
	DisableUnitTests bool
}

func New(
	// top level source code directory
	src *dagger.Directory,
	// .netrc file for private modules can be passed as env var or file --netrc env:var_name, file:/filepath/.netrc
	// +optional
	netrc *dagger.Secret,
) (*Release, error) {

	return &Release{
		Source:         src,
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
