package main

import (
	"context"
	"dagger/python/internal/dagger"
)

// create function test service
func (python *Python) Service(ctx context.Context) *dagger.Service {
	// Run app as a service for function test
	return python.Container().
		WithExposedPort(9333).
		AsService(dagger.ContainerAsServiceOpts{Args: []string{"uv", "run", "start"}})
}

// Return the result of running function test
func (python *Python) FunctionTest(ctx context.Context,
	// function test directory
	// +optional
	// +default="ftest"
	dir string,
) (string, error) {
	functionTest := python.Container().
		WithServiceBinding("localhost", python.Service(ctx)).
		WithExec([]string{"uv", "run", "pytest", dir})

	// Return the formatted output of the function test as a string
	return functionTest.Stdout(ctx)
}
