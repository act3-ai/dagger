package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
)

// run commands with mypy
type Mypy struct {
	// +private
	Python *Python
}
type MypyResults struct {
	// returns results of mypy check as a file
	Results *dagger.File
	// returns exit code of mypy check
	ExitCode int
}

// Runs mypy on a given source directory. Returns a results file and an exit-code.
func (m *Mypy) Check(ctx context.Context,
	// +optional
	outputFormat string,
) (*MypyResults, error) {
	args := []string{
		"uv",
		"run",
		"--with=mypy",
		"mypy",
	}

	// Append outputFormat only if it's provided
	if outputFormat != "" {
		args = append(args, "--output", outputFormat)
	}

	// Add path
	args = append(args, ".")

	ctr, err := m.Python.Container().
		WithExec(args, dagger.ContainerWithExecOpts{
			RedirectStdout: "/mypy-results.txt",
			Expect:         dagger.ReturnTypeAny}).Sync(ctx)

	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running mypy: %w", err)
	}

	results := ctr.File("/mypy-results.txt")

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	return &MypyResults{
		Results:  results,
		ExitCode: exitCode,
	}, nil

}
