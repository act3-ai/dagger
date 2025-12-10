package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
)

type Pylint struct {
	// +private
	Python *Python
}
type PylintResults struct {
	// prints the combined output of stdout and stderr as a string
	// +private
	Output string
	// returns results of pylint as a file
	Results *dagger.File
	// returns exit code of pylint
	ExitCode int
}

// Runs pylint on a given source directory. Returns a results file and an exit-code.
func (p *Python) Pylint(ctx context.Context,
	// +optional
	// +default="text"
	outputFormat string,
) (*PylintResults, error) {

	ctr, err := p.Container().
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
				Expect: dagger.ReturnTypeAny}).
		Sync(ctx)

	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running pylint: %w", err)
	}

	output, err := ctr.CombinedOutput(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	return &PylintResults{
		Output:   output,
		Results:  dag.File("pylint-results.txt", output),
		ExitCode: exitCode,
	}, nil
}

// Check for any errors running pylint
func (pl *PylintResults) Check(ctx context.Context,
) error {
	if pl.ExitCode == 0 {
		return nil
	}

	return fmt.Errorf("%s", pl.Output)
}
