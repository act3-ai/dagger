package main

import (
	"context"
	"dagger/release/internal/dagger"
	"dagger/release/util"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

// TODO: consider adding the release string formmatter to release struct itself
// TODO: helm chart version bumping, make it flexible to zero or more helm charts
// TODO: add support for modifications to releases.md for images and helm chart table
// TODO: unit test specific image base plumbing
// TODO: generate image base plumbing

// Generate release notes, changelog, and target release version.
func (r *Release) Prepare(ctx context.Context,
	// Changelog file path, relative to source directory
	// +optional
	// +default="CHANGELOG.md"
	changelog string,
	// Release notes file path, relative to source directory. Default: releases/v<version>.md.
	// +optional
	notesPath string,
	// base image for git-cliff
	// +optional
	base *dagger.Container,
) (*dagger.Directory, error) {
	targetVersion, err := r.version(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("resolving next release versin: %w", err)
	}

	notesDir := "releases"
	notesName := targetVersion + ".md"
	if notesPath != "" {
		notesDir = filepath.Dir(notesPath)
		notesName = filepath.Base(notesPath)
	}

	releaseNotesFile, err := r.notes(ctx, notesName, base)
	if err != nil {
		return nil, fmt.Errorf("generating release notes: %w", err)
	}
	changelogFile := r.changelog(ctx, changelog, base)

	return dag.Directory().
		WithFile(changelog, changelogFile).
		WithFile(filepath.Join(notesDir, notesName), releaseNotesFile).
		WithNewFile("VERSION", strings.TrimPrefix(targetVersion+"\n", "v")), nil

}

// genericLint runs geneic linters, e.g. markdown, yaml, etc.
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
			if r == "" {
				r = "No reported issues."
			}
			r = fmt.Sprintf("Results for file %s:\n%s", entry, r)
			return r, err
		})
	}

	res, err := p.Wait()
	return strings.Join(res, "\n\n"), err
}

// gitStatus returns an error if a git repository contains uncommitted changes.
func (r *Release) gitStatus(ctx context.Context) error {
	ctr := dag.Wolfi().
		Container(
			dagger.WolfiContainerOpts{
				Packages: []string{"git"},
			},
		).
		WithMountedDirectory("/work/src", r.Source).
		WithWorkdir("/work/src")

	var errs []error
	// check for unstaged changes
	_, err := ctr.WithExec([]string{"git", "diff", "--stat", "--exit-code"}, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}).
		Stdout(ctx)

	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		result := fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr)
		// exit code != 0
		errs = append(errs, fmt.Errorf("checking for unstaged git changes: %s", result))
	case err != nil:
		// some other dagger error, e.g. graphql
		return err
	}

	// check for staged, but not committed changes
	_, err = ctr.WithExec([]string{"git", "diff", "--cached", "--stat", "--exit-code"}, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}).
		Stdout(ctx)
	switch {
	case errors.As(err, &e):
		result := fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr)
		// exit code != 0
		errs = append(errs, fmt.Errorf("checking for staged git changes: %s", result))
	case err != nil:
		// some other dagger error, e.g. graphql
		return err
	}

	return errors.Join(errs...)
}

// Generate the next version from conventional commit messages (see cliff.toml).
//
// Includes 'v' prefix.
func (r *Release) version(ctx context.Context,
	// base image for git-cliff
	// +optional
	base *dagger.Container,
) (string, error) {
	targetVersion, err := dag.GitCliff(r.Source, dagger.GitCliffOpts{Container: base}).
		BumpedVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving release target version: %w", err)
	}

	return strings.TrimSpace(targetVersion), err
}

// Generate the change log from conventional commit messages.
//
// changelog is a default changelog generated using the git-cliff module. Please use the act3-ai/dagger/git-cliff module directly for custom changelogs.
func (r *Release) changelog(ctx context.Context,
	// Changelog file path, relative to source directory
	// +optional
	// +default="CHANGELOG.md"
	changelog string,
	// base image for git-cliff
	// +optional
	base *dagger.Container,
) *dagger.File {
	// generate and prepend to changelog
	return dag.GitCliff(r.Source, dagger.GitCliffOpts{Container: base}).
		WithBump().
		WithStrip("footer").
		WithUnreleased().
		WithPrepend(changelog).
		Run().
		File(changelog)
}

// Generate release notes.
//
// notes are default release notes generated using the git-cliff module. Please use the act3-ai/dagger/git-cliff module directly for custom release notes.
func (r *Release) notes(ctx context.Context,
	// Custom release notes file name. Default: v<version>.md
	// +optional
	name string,
	// base image for git-cliff
	// +optional
	base *dagger.Container,
) (*dagger.File, error) {
	// generate and export release notes
	notes, err := dag.GitCliff(r.Source, dagger.GitCliffOpts{Container: base}).
		WithBump().
		WithUnreleased().
		WithStrip("all").
		Run().
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("generating release notes: %w", err)
	}

	if name == "" {
		version, err := r.version(ctx, base)
		if err != nil {
			return nil, fmt.Errorf("resolving release version for notes file name: %w", err)
		}
		name = fmt.Sprintf("%s.md", version)
	}

	return dag.Directory().
		WithNewFile(name, notes).
		File(name), nil
}
