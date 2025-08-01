package main

import (
	"context"
	"dagger/release/internal/dagger"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Generate release notes, changelog, and target release version.
func (r *Release) Prepare(ctx context.Context,
	// prepare for a specific version, overrides default bumping configuration, prioritized over method.
	// +optional
	version string,
	// prepare for a specific method/type of release, overrides bumping configuration, ignored if version is specified. Supported values: 'major', 'minor', and 'patch'.
	// +optional
	method string,
	// path to version file
	// +optional
	// +default="VERSION"
	versionPath string,
	// path to helm chart in source directory to bump chart version to release version.
	// +optional
	chartPath string,
	// Changelog file path, relative to source directory
	// +optional
	// +default="CHANGELOG.md"
	changelogPath string,
	// Release notes file path, relative to source directory. Default: releases/v<version>.md.
	// +optional
	notesPath string,
	// Additional information to include in release notes. Injected after header and before commit
	// +optional
	extraNotes string,
	// base image for git-cliff
	// +optional
	base *dagger.Container,
	// ignore git status errors
	// +optional
	ignoreError bool,
	// additional arguments to git-cliff --bumped version
	// +optional
	args []string,
) (*dagger.Directory, error) {
	if !ignoreError {
		if err := r.gitStatus(ctx); err != nil {
			return nil, fmt.Errorf("git repository is dirty, aborting prepare: %w", err)
		}
	}

	// bump version if not specified
	var err error
	if version == "" {
		version, err = r.version(ctx, method, base, args)
		if err != nil {
			return nil, fmt.Errorf("resolving next release version: %w", err)
		}
	}

	if notesPath == "" {
		notesPath = filepath.Join("releases", fmt.Sprintf("%s.md", version))
	}
	notesDir := filepath.Dir(notesPath)
	notesName := filepath.Base(notesPath)

	releaseNotesFile, err := r.notes(ctx, version, notesName, extraNotes, base, args)
	if err != nil {
		return nil, fmt.Errorf("generating release notes: %w", err)
	}

	// update changelog
	changelogFile := r.changelog(ctx, version, changelogPath, base, args)

	// set helm chart version
	var chartFile *dagger.File
	if chartPath != "" {
		chartFile = r.setHelmChartVersion(version, chartPath)
	}

	return dag.Directory().
		WithFile(changelogPath, changelogFile).
		WithFile(filepath.Join(notesDir, notesName), releaseNotesFile).
		WithNewFile(versionPath, strings.TrimPrefix(version+"\n", "v")).
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
		With(func(c *dagger.Container) *dagger.Container {
			if r.GitIgnore != nil {
				const gitIgnorePath = "/work/.gitignore"
				c = c.WithMountedFile(gitIgnorePath, r.GitIgnore).
					WithExec([]string{"git", "config", "--global", "core.excludesfile", gitIgnorePath})
			}
			return c
		}).
		WithWorkdir("/work/src")

	var errs []error
	// check for unstaged changes
	_, err := ctr.WithExec([]string{"git", "diff", "--stat", "--exit-code"}, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeSuccess}).
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
	_, err = ctr.WithExec([]string{"git", "diff", "--cached", "--stat", "--exit-code"}, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeSuccess}).
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
	// prepare for a specific method/type of release, overrides bumping configuration, ignored if version is specified. Supported values: 'major', 'minor', and 'patch'.
	// +optional
	method string,
	// base image for git-cliff
	// +optional
	base *dagger.Container,
	// additional arguments and flags for git-cliff --bumped-version
	// +optional
	args []string,
) (string, error) {
	version, err := dag.GitCliff(r.Source, dagger.GitCliffOpts{Container: base}).
		With(func(r *dagger.GitCliff) *dagger.GitCliff {
			// method="" throws an error
			if method != "" {
				r = r.WithBump(dagger.GitCliffWithBumpOpts{Method: method})
			}
			return r
		}).
		BumpedVersion(ctx, dagger.GitCliffBumpedVersionOpts{Args: args})

	return strings.TrimSpace(version), err
}

// Generate the change log from conventional commit messages.
//
// changelog is a default changelog generated using the git-cliff module. Please use the act3-ai/dagger/git-cliff module directly for custom changelogs.
func (r *Release) changelog(ctx context.Context,
	version string,
	// Changelog file path, relative to source directory
	// +optional
	// +default="CHANGELOG.md"
	changelog string,
	// base image for git-cliff
	// +optional
	base *dagger.Container,
	// additional arguments and flags for git-cliff
	// +optional
	args []string,
) *dagger.File {
	// generate and prepend to changelog
	return dag.GitCliff(r.Source, dagger.GitCliffOpts{Container: base}).
		WithTag(version).
		WithStrip("footer").
		WithUnreleased().
		WithPrepend(changelog).
		Run(dagger.GitCliffRunOpts{Args: args}).
		File(changelog)
}

// Generate release notes.
//
// notes are default release notes generated using the git-cliff module. Please use the act3-ai/dagger/git-cliff module directly for custom release notes.
func (r *Release) notes(ctx context.Context,
	version string,
	// Custom release notes file name. Default: v<version>.md
	// +optional
	name string,
	// Additional information to include in release notes. Injected after header and before commit
	// +optional
	extraNotes string,
	// base image for git-cliff
	// +optional
	base *dagger.Container,
	// additional arguments and flags for git-cliff
	// +optional
	args []string,
) (*dagger.File, error) {
	// generate and export release notes
	notes, err := dag.GitCliff(r.Source, dagger.GitCliffOpts{Container: base}).
		WithTag(version).
		WithUnreleased().
		WithStrip("all").
		Run(dagger.GitCliffRunOpts{Args: args}).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("generating release notes: %w", err)
	}

	// add extra notes section
	if extraNotes != "" {
		b := &strings.Builder{}
		b.WriteString(extraNotes)
		b.WriteString("###")
		notes = strings.Replace(notes, "###", b.String(), 1)
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
