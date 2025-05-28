package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"errors"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

// This file contains generic linters used in 'Check' commands for various project types. Language specific linters are used in their respective groups.

// genericLint runs generic linters, e.g. markdown, yaml, etc.
func (r *Release) genericLint(ctx context.Context,
	results util.ResultsFormatter,
	base *dagger.Container,
) error {
	var errs []error

	// TODO: this module does not support a custom base container.
	res, err := r.shellcheck(ctx, 4) // TODO: plumb concurrency?
	results.Add("Shellcheck", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running shellcheck: %w", err))
	}

	res, err = dag.Yamllint(r.Source, dagger.YamllintOpts{Base: base}).
		Run(ctx)
	results.Add("Yamllint", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running yamllint: %w", err))
	}

	res, err = dag.Markdownlint(r.Source, dagger.MarkdownlintOpts{Base: base}).
		Run(ctx)
	results.Add("Markdownlint", res)
	if err != nil {
		errs = append(errs, fmt.Errorf("running markdownlint: %w", err))
	}

	return errors.Join(errs...)
}

// shellcheck auto-detects and runs on all *.sh and *.bash files in the source directory.
//
// Users who want custom functionality should use github.com/dagger/dagger/modules/shellcheck directly.
func (r *Release) shellcheck(ctx context.Context, concurrency int) (string, error) {

	// TODO: Consider adding an option for specifying script files that don't have the extension, such as WithShellScripts.
	shEntries, err := r.Source.Glob(ctx, "**/*.sh")
	if err != nil {
		return "", fmt.Errorf("globbing shell scripts with *.sh extension: %w", err)
	}

	bashEntries, err := r.Source.Glob(ctx, "**/*.bash")
	if err != nil {
		return "", fmt.Errorf("globbing shell scripts with *.bash extension: %w", err)
	}

	p := pool.NewWithResults[string]().
		WithMaxGoroutines(concurrency).
		WithErrors().
		WithContext(ctx)

	entries := append(shEntries, bashEntries...)
	for _, entry := range entries {
		p.Go(func(ctx context.Context) (string, error) {
			r, err := dag.Shellcheck().
				Check(r.Source.File(entry)).
				Report(ctx)
			// this is needed because of weird error handling  in shellcheck here:
			// https://github.com/dagger/dagger/blob/0b46ea3c49b5d67509f67747742e5d8b24be9ef7/modules/shellcheck/main.go#L137
			if r != "" {
				return "", fmt.Errorf("results for file %s:\n%s", entry, r)
			}
			// r = fmt.Sprintf("Results for file %s:\n%s", entry, r)
			return r, err
		})
	}

	res, err := p.Wait()
	return strings.Join(res, "\n\n"), err
}
