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

	"github.com/dagger/dagger/util/parallel"
)

type Tests struct{}

const ref = "registry:5000/test/acedt:v1"

// +check
// Run all tests.
func (t *Tests) All(ctx context.Context) error {

	if err := t.Gather(ctx); err != nil {
		return err
	}
	return parallel.New().WithLimit(3).
		WithJob("Scatter", t.Scatter).
		WithJob("serialize", t.Serialize).
		WithJob("Archive", t.Archive).
		Run(ctx)
}

func (t *Tests) RunSvc(ctx context.Context) *dagger.Service {

	return dag.Container().
		From("docker.io/library/registry:3.0.0-rc.3").
		WithMountedCache("/var/lib/registry", dag.CacheVolume("docker-registry")).
		WithExposedPort(5000).
		AsService()

}

// Run test for Gather
func (t *Tests) Gather(ctx context.Context) error {
	src := dag.CurrentModule().Source()
	config := src.File("testdata/config.yaml")
	artifacts := src.File("testdata/artifacts.csv")

	registry := t.RunSvc(ctx)

	c := dag.DataTool().Container()
	c = c.WithServiceBinding("registry", registry).WithFile("/root/.config/ace/dt/config.yaml", config)

	_, err := dag.DataTool(dagger.DataToolOpts{Base: c}).Gather(ctx, artifacts, ref)
	if err != nil {
		return err
	}

	return err

}

// Run test for Scatter
func (t *Tests) Scatter(ctx context.Context) error {
	src := dag.CurrentModule().Source()
	config := src.File("testdata/config.yaml")
	mapping := src.File("testdata/mapping.csv")

	registry := t.RunSvc(ctx)

	c := dag.DataTool().Container()
	c = c.WithServiceBinding("registry", registry).WithFile("/root/.config/ace/dt/config.yaml", config)

	err := dag.DataTool(dagger.DataToolOpts{Base: c}).Scatter(ctx, ref, mapping)

	return err

}

// Run test for Serialize
func (t *Tests) Serialize(ctx context.Context) error {
	src := dag.CurrentModule().Source()
	config := src.File("testdata/config.yaml")

	registry := t.RunSvc(ctx)

	c := dag.DataTool().Container()
	c = c.WithServiceBinding("registry", registry).WithFile("/root/.config/ace/dt/config.yaml", config)

	_, err := dag.DataTool(dagger.DataToolOpts{Base: c}).Serialize(ref).Name(ctx)

	return err

}

// Run test for Archive
func (t *Tests) Archive(ctx context.Context) error {
	src := dag.CurrentModule().Source()
	config := src.File("testdata/config.yaml")
	artifacts := src.File("testdata/artifacts.csv")

	c := dag.DataTool().Container()
	c = c.WithFile("/root/.config/ace/dt/config.yaml", config)

	_, err := dag.DataTool(dagger.DataToolOpts{Base: c}).Archive(artifacts).Name(ctx)

	return err

}
