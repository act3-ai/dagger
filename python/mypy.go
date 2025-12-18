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
	// +private
	ExitCode int
}

// Runs mypy on a given source directory.
func (p *Python) Mypy(ctx context.Context,
	// +optional
	outputFormat string) (*MypyResults, error) {
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

	ctr, err := p.Container().
		WithExec(args, dagger.ContainerWithExecOpts{
			Expect: dagger.ReturnTypeAny}).Sync(ctx)

	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running mypy: %w", err)
	}

	output, err := ctr.CombinedOutput(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("getting results: %w", err)
	}

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("getting exit code: %w", err)
	}
	return &MypyResults{
		Results:  dag.File("mypy-results.txt", output),
		ExitCode: exitCode,
	}, nil
}

// Check for any errors running mypy
func (mr *MypyResults) Check(ctx context.Context) error {
	if mr.ExitCode == 0 {
		return nil
	}
	results, err := mr.Results.Contents(ctx)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", results)
}
