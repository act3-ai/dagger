// A generated module for Kubeconform functions
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
	"dagger/kubeconform/internal/dagger"
	"fmt"
	"strings"
)

type Kubeconform struct{}

// validates a directory of Kubernetes manifests against standard and CRD schemas.
func (m *Kubeconform) Validate(ctx context.Context,
	// directory of kubernetes manifests to validate
	manifests *dagger.Directory,
	// Directory containing CRD definitions to validate with.
	// Must be in .yaml format
	// +optional
	crds *dagger.Directory,
) *dagger.Container {
	validateCtr := m.bashCtr().
		WithDirectory("/manifests", manifests)

	args := []string{
		"/kubeconform", "-strict",
		"-schema-location", "default",
	}

	// convert crds to openapi2 schemas and validate with them if provided
	if crds != nil {
		schemasDir := m.convertCrdsToSchemas(crds)
		validateCtr = validateCtr.WithDirectory("/schemas", schemasDir)
		args = append(args, "-schema-location", "/schemas/{{ .ResourceKind }}-{{ .Group }}-{{ .ResourceAPIVersion }}.json")
	}

	args = append(args,
		"-schema-location",
		"https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json",
		"/manifests/.",
	)

	return validateCtr.
		WithExec(args)
}

// ValidateKustomize builds Kustomize directories and validates the resulting manifests.
func (m *Kubeconform) ValidateKustomize(
	ctx context.Context,
	// Top-level source directory
	// +defaultPath="/"
	src *dagger.Directory,
	// Relative paths containing kustomization files to build and validate
	// ex. --paths="overlays/dev,overlays/prod"
	paths []string,
	// Optional directory of CRDs to validate with
	// +optional
	crds *dagger.Directory,
) *dagger.Container {
	manifests := m.KustomizeBuild(src, paths)
	return m.Validate(ctx, manifests, crds)
}

// builds rendered Kubernetes manifests for specified paths using `kubectl kustomize`.
// Returns a directory containing the rendered .yaml files.
func (m *Kubeconform) KustomizeBuild(
	// top level source code directory
	// +defaultPath="/"
	src *dagger.Directory,
	// Relative paths containing kustomization files to build and validate
	// ex. --paths="overlays/dev,overlays/prod"
	paths []string) *dagger.Directory {

	ctr := dag.Container().
		From("bitnami/kubectl").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithDirectory("/out", dag.Directory())

	for _, p := range paths {
		outFile := fmt.Sprintf("/out/%s.yaml", strings.ReplaceAll(p, "/", "_"))
		ctr = ctr.WithExec([]string{"kubectl", "kustomize", p}, dagger.ContainerWithExecOpts{
			RedirectStdout: outFile,
		})
	}

	return ctr.Directory("/out")

}

// helper container so that bash is accessible
func (m *Kubeconform) bashCtr() *dagger.Container {

	kcCtr := dag.Container().
		From("ghcr.io/yannh/kubeconform:latest")

	return dag.Container().From("cgr.dev/chainguard/bash").
		WithFile("/kubeconform", kcCtr.File("/kubeconform")).
		WithFile("/openapi2jsonschema", kcCtr.File("/openapi2jsonschema")).
		WithEnvVariable("FILENAME_FORMAT", "{kind}-{fullgroup}-{version}")

}

// helper to generate openapi2jsonschemas from a crd
func (m *Kubeconform) convertCrdsToSchemas(crds *dagger.Directory) *dagger.Directory {
	return m.bashCtr().
		WithWorkdir("/schemas").
		WithDirectory("/crds/", crds).
		// converts all crds to openapi2jsonschema
		WithExec([]string{"/bin/bash", "-c", "/openapi2jsonschema /crds/*.yaml"}).
		// return the generated JSON schemas
		Directory("/schemas")
}
