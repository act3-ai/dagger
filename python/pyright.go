package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
)

type Pyright struct {
	// +private
	Python *Python
}
type PyrightResults struct {
	// returns results of pyright as a file
	Results *dagger.File
	// returns exit code of pyright
	// +private
	ExitCode int
}

// Runs pyright on a given source directory. Returns a results file and an exit-code.
func (p *Python) Pyright(ctx context.Context,
) (*PyrightResults, error) {

	ctr, err := p.Container().
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=pyright",
				"pyright",
				".",
			}, dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny}).Sync(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running pyright: %w", err)
	}

	output, err := ctr.CombinedOutput(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get results: %w", err)
	}

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	return &PyrightResults{
		Results:  dag.File("pyright-results.txt", output),
		ExitCode: exitCode,
	}, nil

}

// Check for any errors running pyright
func (pr *PyrightResults) Check(ctx context.Context,
) error {
	if pr.ExitCode == 0 {
		return nil
	}
	results, err := pr.Results.Contents(ctx)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", results)
}
