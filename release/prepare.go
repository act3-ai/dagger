package main

import (
	"context"
	"dagger/release/internal/dagger"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// TODO: consider adding the release string formmatter to release struct itself
// TODO: helm chart version bumping, make it flexible to zero or more helm charts
// TODO: add support for modifications to releases.md for images and helm chart table
// TODO: unit test specific image base plumbing
// TODO: generate image base plumbing

// Generate release notes, changelog, and target release version.
func (r *Release) Prepare(ctx context.Context,
	// path to helm chart in source directory to bump chart version to release version.
	// +optional
	chartPath string,
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
	if err := r.gitStatus(ctx); err != nil {
		return nil, fmt.Errorf("git repository is dirty, aborting prepare: %w", err)
	}

	// update version file
	targetVersion, err := r.version(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("resolving next release versin: %w", err)
	}

	// update release notes file
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

	// update changelog
	changelogFile := r.changelog(ctx, changelog, base)

	// set helm chart version
	var chartFile *dagger.File
	if chartPath != "" {
		chartFile = r.setHelmChartVersion(targetVersion, chartPath)
	}

	return dag.Directory().
		WithFile(changelog, changelogFile).
		WithFile(filepath.Join(notesDir, notesName), releaseNotesFile).
		WithNewFile("VERSION", strings.TrimPrefix(targetVersion+"\n", "v")).
		With(func(d *dagger.Directory) *dagger.Directory {
			if chartFile != nil {
				d = d.WithFile(chartPath, chartFile)
			}
			return d
		}), nil
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

// Set the version and appVersion of a helm chart.
func (r *Release) setHelmChartVersion(
	// release version
	version string,
	// Chart.yaml path
	chartPath string,
) *dagger.File {
	version = strings.TrimPrefix(version, "v")
	updatedChart := dag.Wolfi().
		Container(dagger.WolfiContainerOpts{
			Packages: []string{"yq"},
		}).
		WithMountedDirectory("/src", r.Source).
		WithWorkdir("/src").
		WithEnvVariable("version", version).
		WithExec([]string{"yq", "e",
			"(.version = env(version)) | (.appVersion = \"v\"+env(version))",
			"-i", chartPath}).
		File(chartPath)

	return updatedChart
}
