package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"fmt"
	"slices"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

// This file contains generic linters used in 'Check' commands for various project types. Language specific linters are used in their respective groups.

// GenericLint runs shellcheck, yamllint, and markdownlint
func (r *Release) GenericLint(ctx context.Context,
	// +optional
	base *dagger.Container,
	// skip any provided lint tests
	// +optional
	skip []string,
) (string, error) {
	results := util.NewResultsBasicFmt(strings.Repeat("=", 15))

	p := pool.New().
		WithErrors().
		WithContext(ctx)

	if !slices.Contains(skip, "shellcheck") {
		p.Go(func(ctx context.Context) error {
			res, err := r.shellcheck(ctx, 4) // TODO: plumb concurrency?
			results.Add("Shellcheck", res)
			if err != nil {
				return fmt.Errorf("running shellcheck: %w", err)
			}
			return nil
		})
	}

	if !slices.Contains(skip, "yamllint") {
		p.Go(func(ctx context.Context) error {
			res, err := dag.Yamllint(r.Source, dagger.YamllintOpts{Base: base}).
				Run(ctx)
			results.Add("Yamllint", res)
			if err != nil {
				return fmt.Errorf("running yamllint: %w", err)
			}
			return nil
		})
	}

	if !slices.Contains(skip, "markdownlint") {
		p.Go(func(ctx context.Context) error {
			res, err := dag.Markdownlint(r.Source, dagger.MarkdownlintOpts{Base: base}).
				Run(ctx)
			results.Add("Markdownlint", res)
			if err != nil {
				return fmt.Errorf("running markdownlint: %w", err)
			}
			return nil
		})
	}

	if err := p.Wait(); err != nil {
		return "", err
	}

	return results.String(), nil
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
