// A generated module for Tests functions
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
	"dagger/tests/internal/dagger"
)

type Tests struct{}

// +check
func (m *Tests) Manifest() *dagger.Container {
	manifest := dag.CurrentModule().Source().
		Directory("testdata/").
		Filter(dagger.DirectoryFilterOpts{Include: []string{"kustomize/manifest.yaml"}})

	return dag.Kubeconform().Validate(manifest, dagger.KubeconformValidateOpts{Verbose: true})
}

// +check
func (m *Tests) KustomizeWithCrds() *dagger.Container {
	manifest := dag.CurrentModule().Source().Directory("testdata")
	return dag.Kubeconform().
		ValidateKustomize(manifest, []string{"kustomize"}, dagger.KubeconformValidateKustomizeOpts{Crds: manifest.Directory("crds")})
}
