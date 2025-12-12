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
	"context"
	"dagger/tests/internal/dagger"
	"fmt"
	"strings"
)

type Tests struct{}

// +check
// Test WithLabels to ensure name and value are being set properly
func (t *Tests) WithLabels(ctx context.Context,
	// +defaultPath="."
	src *dagger.Directory) error {

	const expectedName = `test.io`
	const expectedValue = `label1`

	ctr := dag.Docker(src).WithLabel("test.io", "label1").
		Build(dagger.DockerBuildOpts{Target: "with-label"})

	labelList, err := ctr.Labels(ctx)
	if err != nil {
		return err
	}

	for _, label := range labelList {
		actualName, err := label.Name(ctx)
		if err != nil {
			return err
		}
		if strings.TrimSpace(actualName) != expectedName {
			return fmt.Errorf("label name does not match the expected value\nactual:   %s\nexpected: %s", actualName, expectedName)
		}
		actualValue, err := label.Value(ctx)
		if err != nil {
			return err
		}
		if strings.TrimSpace(actualValue) != expectedValue {
			return fmt.Errorf("label value does not match the expected value\nactual:   %s\nexpected: %s", actualValue, expectedValue)
		}
	}

	return err

}

// +check
func (t *Tests) WithBuildArg(ctx context.Context,
	// +defaultPath="."
	src *dagger.Directory) error {

	// dagger has no list build arg function, so build arg is being set to file in Dockerfile
	// .Stdout() requires a WithExec, even if one is in Dockerfile
	actual, err := dag.Docker(src).WithBuildArg("TEST_ARG1", "testvalue1").
		Build(dagger.DockerBuildOpts{Target: "with-build-arg"}).WithExec([]string{"cat", "testarg.txt"}).Stdout(ctx)

	const expected = "testvalue1"

	if strings.TrimSpace(actual) != expected {
		return fmt.Errorf("build arg value does not match the expected value\nactual:   %s\nexpected: %s", actual, expected)
	}

	return err

}

// + check
// Test WithSecrets to ensure name and value are being set properly
func (t *Tests) WithSecret(ctx context.Context,
	// +defaultPath="."
	src *dagger.Directory) error {

	_, err := dag.Docker(src).
		WithSecret("TEST_SECRET1", dag.SetSecret("DUMMY_SECRET1", "password1")).
		Build(dagger.DockerBuildOpts{Target: "with-secret"}).
		Sync(ctx)

	return err

}

// tests for the future when publish is actually testable
// Test WithRegistryAuth to ensure name and value are being set properly
// func (t *Tests) WithRegistryAuth(ctx context.Context,
// 	// +defaultPath="."
// 	src *dagger.Directory,
// 	svc *dagger.Service) error {

// 	image := dag.Docker(dagger.DockerOpts{Src: src}).
// 		Build(dagger.DockerBuildOpts{Target: "with-label"}).
// 		AsTarball()

// 	// registry := t.RunSvc(ctx)

// 	_, err := dag.Container().From("quay.io/skopeo/stable").
// 		WithServiceBinding("registry", svc).
// 		WithMountedFile("/work/image.tar", image).
// 		WithExec([]string{"skopeo", "copy", "--all", "--dest-tls-verify=false", "docker-archive:/work/image.tar", "docker://registry:5000/test/test:v1.0.0"}).
// 		Sync(ctx)

// 	return err

// }

// // Dagger cannot use internal services for publish command
// // issue: https://github.com/dagger/dagger/issues/6411
// func (t *Tests) RunSvc(ctx context.Context) *dagger.Service {

// 	return dag.Container().
// 		From("docker.io/library/registry:3.0.0-rc.3").
// 		WithExposedPort(5000).
// 		AsService()

// }
