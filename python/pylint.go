package main

import (
	"context"
	"dagger/python/internal/dagger"
	"errors"
	"fmt"
)

// Return the result of running pylint
func (python *Python) Pylint(ctx context.Context,
	// +optional
	// +default="text"
	outputFormat string,
	// ignore errors and return result
	// +optional
	ignoreError bool) (string, error) {

	out, err := python.Container().
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
				".",
			}).Stdout(ctx)

	var e *dagger.ExecError

	switch {
	case errors.As(err, &e):
		if ignoreError {
			return fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr), nil
		}
		return "", fmt.Errorf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr)
	case err != nil:
		// some other dagger error, e.g. graphql
		return "", fmt.Errorf("Stout:\n%w", err)
	default:
		// stdout of the linter with exit code 0
		return out, nil
	}

}
