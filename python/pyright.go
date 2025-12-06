package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
)

type PyRightResults struct {
	// returns results of pyright as a file
	Results *dagger.File
	// returns exit code of pyright
	ExitCode int
}

// +check
// Return the result of running Pyright
func (python *Python) Pyright(ctx context.Context,
) (*PyRightResults, error) {

	ctr, err := python.Container().
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=pyright",
				"pyright",
				".",
			}, dagger.ContainerWithExecOpts{
				RedirectStdout: "/pyright-results.txt",
				Expect:         dagger.ReturnTypeAny}).Sync(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running pyright: %w", err)
	}

	results := ctr.File("/pyright-results.txt")

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	return &PyRightResults{
		Results:  results,
		ExitCode: exitCode,
	}, nil

}
