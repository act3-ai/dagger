package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
)

type PylintResults struct {
	// returns results of pylint as a file
	Results *dagger.File
	// returns exit code of pylint
	ExitCode int
}

// Return the result of running pylint
func (python *Python) Pylint(ctx context.Context,
	// +optional
	// +default="text"
	outputFormat string,
) (*PylintResults, error) {

	ctr, err := python.Container().
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=pylint",
				"pylint",
				"--recursive=y",
				"--persistent=n",
				"--ignore-paths=.venv",
				"--output-format", outputFormat,
				// "--reports=y",
				"."},
			dagger.ContainerWithExecOpts{
				RedirectStdout: "/pylint-results.txt",
				Expect:         dagger.ReturnTypeAny}).
		Sync(ctx)

	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running pylint: %w", err)
	}

	results := ctr.File("/pylint-results.txt")

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	return &PylintResults{
		Results:  results,
		ExitCode: exitCode,
	}, nil
}
