package main

import (
	"context"
	"dagger/python/internal/dagger"
	"errors"
	"fmt"
)

// Return the result of running ruff check
func (python *Python) RuffCheck(ctx context.Context,
	// +optional
	// +default="full"
	outputFormat string,
	// +optional
	ignoreError bool,
) (string, error) {
	// Run the Ruff linter with the provided output format
	out, err := python.Container().WithExec(
		[]string{
			"uv",
			"run",
			"--with=ruff",
			"ruff",
			"check", ".",
			"--output-format", outputFormat}).Stdout(ctx)
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

// Return the result of running ruff format
func (python *Python) RuffFormat(ctx context.Context,
	// ignore errors and return result
	// +optional
	ignoreError bool) (string, error) {

	out, err := python.Container().
		WithExec(
			[]string{
				"uv",
				"run",
				"--with=ruff",
				"ruff",
				"format",
				"--check",
				"--diff", "."}).Stdout(ctx)

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
