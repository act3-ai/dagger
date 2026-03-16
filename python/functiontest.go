package main

import (
	"context"
	"dagger/python/internal/dagger"
)

// create function test service
func (python *Python) Service(ctx context.Context) (*dagger.Service, error) {
	ctr, err := python.Runtime(ctx)
	if err != nil {
		return nil, err
	}
	// Run app as a service for function test
	return ctr.
		WithExposedPort(9333).
		AsService(dagger.ContainerAsServiceOpts{Args: []string{"uv", "run", "start"}}), nil
}

// Return the result of running function test
func (python *Python) FunctionTest(ctx context.Context,
	// function test directory
	// +optional
	// +default="ftest"
	dir string,
) (string, error) {
	ctr, err := python.Runtime(ctx)
	if err != nil {
		return "", err
	}
	svc, err := python.Service(ctx)
	if err != nil {
		return "", err
	}
	functionTest := ctr.
		WithServiceBinding("localhost", svc).
		WithExec([]string{"uv", "run", "pytest", dir})

	// Return the formatted output of the function test as a string
	return functionTest.Stdout(ctx)
}
