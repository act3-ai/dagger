package main

import (
	"context"
	"dagger/act-3-pipeline/internal/dagger"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/sourcegraph/conc/pool"
)

// This file contains generic linters used in 'Check' commands for various project types. Language specific linters are used in their respective groups.

type Linter struct {
	Name    string
	Feature string
	Run     func(ctx context.Context, src *dagger.Directory) (string, error)
}

// Helper: parse skips into a map for quick lookup
func newSkipSet(skips []string) map[string]bool {
	set := make(map[string]bool)
	for _, s := range skips {
		set[strings.ToLower(strings.TrimSpace(s))] = true
	}
	return set
}

// Helper: Generate list of default linters. This is needed due to args not being present until runtime.
func getDefaultLinters(a *Act3Pipeline, base *dagger.Container) []Linter {
	return []Linter{
		{
			Name:    "markdownlint",
			Feature: "markdown",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {
				return dag.Markdownlint(a.Source, dagger.MarkdownlintOpts{Base: base}).Run(ctx)
			},
		},
		{
			Name:    "yamllint",
			Feature: "yaml",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {
				return dag.Yamllint(a.Source, dagger.YamllintOpts{Base: base}).Run(ctx)
			},
		},
		{
			Name:    "shellcheck",
			Feature: "shell",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {

				return a.shellcheck(ctx, 4)
			},
		},
		{
			Name:    "ruff-check",
			Feature: "python",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {

				return dag.Python(a.Source).RuffCheck(ctx)
			},
		},
		{
			Name:    "ruff-format",
			Feature: "python",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {

				return dag.Python(a.Source).RuffFormat(ctx)
			},
		},
		{
			Name:    "mypy",
			Feature: "python",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {

				return dag.Python(a.Source).Mypy(ctx)
			},
		},
		{
			Name:    "pylint",
			Feature: "python",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {

				return dag.Python(a.Source).Pylint(ctx)
			},
		},
		{
			Name:    "pyright",
			Feature: "python",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {

				return dag.Python(a.Source).Pyright(ctx)
			},
		},
		{
			Name:    "golangcilint",
			Feature: "golang",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {

				return dag.GolangciLint(dagger.GolangciLintOpts{Container: base}).
					Run(a.Source, dagger.GolangciLintRunOpts{Timeout: "10m"}).
					Stdout(ctx)
			},
		},
		{
			Name:    "govulncheck",
			Feature: "golang",
			Run: func(ctx context.Context, src *dagger.Directory) (string, error) {

				return dag.Govulncheck(
					dagger.GovulncheckOpts{
						Container: base,
						Netrc:     a.Netrc,
					}).
					ScanSource(ctx, a.Source)
			},
		},
	}
}

// Detect features based on file presence
func DetectFeatures(ctx context.Context, source *dagger.Directory) (map[string]bool, error) {
	m := map[string]bool{}

	// Markdown files
	mdMatches, err := source.Glob(ctx, "**/*.md")
	if err != nil {
		return nil, fmt.Errorf("error detecting markdown files: %w", err)
	}
	m["markdown"] = (len(mdMatches) > 0)

	// yaml files
	yamlMatches, err := source.Glob(ctx, "**/*.y*ml")
	if err != nil {
		return nil, fmt.Errorf("error detecting yaml files: %w", err)
	}
	m["yaml"] = (len(yamlMatches) > 0)

	// bash files
	shellMatches, err := source.Glob(ctx, "**/*.sh")
	if err != nil {
		return nil, fmt.Errorf("error detecting shellcheck files: %w", err)
	}

	m["shell"] = (len(shellMatches) > 0)

	// python files
	pythonMatches, err := source.Glob(ctx, "**/*.py")
	if err != nil {
		return nil, fmt.Errorf("error detecting python files: %w", err)
	}

	m["python"] = (len(pythonMatches) > 0)

	// golang files
	golangMatches, err := source.Glob(ctx, "go.mod")
	if err != nil {
		return nil, fmt.Errorf("error detecting golang files: %w", err)
	}

	m["golang"] = (len(golangMatches) > 0)
	// //ruff-check
	// ruffCheckMatches, err := source.Glob(ctx, "**/*.py")
	// if err != nil {
	// 	return nil, fmt.Errorf("error detecting python files: %w", err)
	// }

	// m["ruff-check"] = (len(ruffCheckMatches) > 0)

	// //ruff-format
	// ruffFormatMatches, err := source.Glob(ctx, "**/*.py")
	// if err != nil {
	// 	return nil, fmt.Errorf("error detecting python files: %w", err)
	// }

	// m["shellcheck"] = (len(ruffFormatMatches) > 0)

	// //mypy
	// mypyMatches, err := source.Glob(ctx, "**/*.py")
	// if err != nil {
	// 	return nil, fmt.Errorf("error detecting python files: %w", err)
	// }

	// m["mypy"] = (len(mypyMatches) > 0)

	// //pylint
	// pylintMatches, err := source.Glob(ctx, "**/*.py")
	// if err != nil {
	// 	return nil, fmt.Errorf("error detecting python files: %w", err)
	// }

	// m["pylint"] = (len(pylintMatches) > 0)

	// //pyright
	// pyrightMatches, err := source.Glob(ctx, "**/*.py")
	// if err != nil {
	// 	return nil, fmt.Errorf("error detecting python files: %w", err)
	// }

	// m["pyright"] = (len(pyrightMatches) > 0)

	return m, nil
}

// helper: filters linters by features detected in project source repo/directory
func filterFeatureMatchedLinters(
	linters []Linter,
	features map[string]bool,
) []Linter {
	var matched []Linter
	for _, l := range linters {
		if features[l.Feature] {
			matched = append(matched, l)
		}
	}
	return matched
}

// helper: filters linters by user skipped linters
func filterSkippedLinters(
	linters []Linter,
	skips map[string]bool,
	results *strings.Builder,
) []Linter {
	var applicable []Linter
	for _, l := range linters {
		if skips[l.Name] {
			results.WriteString(fmt.Sprintf("✖ %s — Skipped (user request)\n", l.Name))
			continue
		}
		applicable = append(applicable, l)
	}
	return applicable
}

// Runs applicable linters based on detected project features in a source directory
func (a *Act3Pipeline) Lint(ctx context.Context,
	//	Source *dagger.Directory,
	// +optional
	base *dagger.Container,
	// +optional
	skip []string,
) (string, error) {
	results := &strings.Builder{}
	skips := newSkipSet(skip)

	// detect project features for appropriate linters to run
	features, err := DetectFeatures(ctx, a.Source)
	if err != nil {
		return "", err
	}

	defaultLinters := getDefaultLinters(a, base)

	// filter applicable linters by feature
	featureLinters := filterFeatureMatchedLinters(defaultLinters, features)

	// filter by skip flags
	applicableLinters := filterSkippedLinters(featureLinters, skips, results)

	if len(applicableLinters) == 0 {
		results.WriteString("No applicable linters to run.\n")
		return results.String(), nil
	}

	// range through applicable linters and run them
	var failed atomic.Bool

	pl := pool.New().WithErrors().WithContext(ctx)

	for _, l := range applicableLinters {

		pl.Go(func(ctx context.Context) error {
			out, err := l.Run(ctx, a.Source)
			if err != nil {
				failed.Store(true)
				results.WriteString(fmt.Sprintf("✖ %s — Failed: %v\n%s\n", l.Name, err, out))
			} else {
				results.WriteString(fmt.Sprintf("✔ %s — Success\n%s\n", l.Name, out))
			}
			return err
		})
	}

	_ = pl.Wait()

	if failed.Load() {
		return "", fmt.Errorf(results.String())
	} else {
		results.WriteString("\n✅ All linters passed successfully.\n")
	}

	return results.String(), nil
}

// shellcheck auto-detects and runs on all *.sh and *.bash files in the source directory.
//
// Users who want custom functionality should use github.com/dagger/dagger/modules/shellcheck directly.
func (a *Act3Pipeline) shellcheck(ctx context.Context, concurrency int) (string, error) {

	// TODO: Consider adding an option for specifying script files that don't have the extension, such as WithShellScripts.
	shEntries, err := a.Source.Glob(ctx, "**/*.sh")
	if err != nil {
		return "", fmt.Errorf("globbing shell scripts with *.sh extension: %w", err)
	}

	bashEntries, err := a.Source.Glob(ctx, "**/*.bash")
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
				Check(a.Source.File(entry)).
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
	return strings.Join(res, ""), err
}
