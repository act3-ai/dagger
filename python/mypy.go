package main

import (
	"context"
	"dagger/python/internal/dagger"
	"errors"
	"fmt"
)

// Return the result of running mypy
func (python *Python) Mypy(ctx context.Context,
	// +optional
	outputFormat string,
	// ignore errors and return result
	// +optional
	ignoreError bool) (string, error) {

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

	out, err := python.Container().
		WithExec(args).Stdout(ctx)

	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		if ignoreError {
			return fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr), nil
		}
		return "", fmt.Errorf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr)
	case err != nil:
		// some other dagger error, e.g. graphql
		return "", err
	default:
		// stdout of the linter with exit code 0
		return out, nil
	}

}
