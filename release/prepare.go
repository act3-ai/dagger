package main

import (
	"context"
	"dagger/release/internal/dagger"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Generate release notes, changelog, and VERSION file with target release version.
// Will also optionally bump a version in provided helm chart path.
func (r *Release) Prepare(ctx context.Context,
	// prepare for a specific version
	version string,
	// prefix to path for changelog, version, and release notes
	// +optional
	pathPrefix string,
	// path to helm chart in source directory to bump chart version to release version.
	// +optional
	chartPath string,
	// Additional information to include in release notes. Injected after header and before commit
	// +optional
	extraNotes string,
	//git-cliff cliff.toml path to use. Defaults to root of gitref "."
	// +optional
	config string,
	//provide a github token to git-cliff.
	//This is needed to avoid github api rate limit
	// +optional
	token *dagger.Secret,
) (*dagger.Changeset, error) {
	version = strings.TrimPrefix(version, "v")
	src := r.GitRef.Tree()

	// Base paths
	changelogPath := "CHANGELOG.md"
	notesPath := filepath.Join("releases", fmt.Sprintf("v%s.md", version))
	versionPath := "VERSION"

	// Prepend path prefix if given
	if pathPrefix != "" {
		changelogPath = filepath.Join(pathPrefix, changelogPath)
		notesPath = filepath.Join(pathPrefix, notesPath)
		versionPath = filepath.Join(pathPrefix, versionPath)
	}
	// generate changelog
	changelogFile := r.changelog(ctx, version, changelogPath, config, token)
	// generate release notes
	releaseNotesFile, err := r.notes(ctx, version, filepath.Base(notesPath), extraNotes, config, token)
	if err != nil {
		return nil, fmt.Errorf("generating release notes: %w", err)
	}

	// set helm chart version
	var chartFile *dagger.File
	if chartPath != "" {
		chartFile = r.setHelmChartVersion(version, chartPath)
	}

	// consider changing the construction of this diff
	// instead just modify the source directory directly and then compute the changes
	after := src.
		WithFile(changelogPath, changelogFile).
		WithFile(notesPath, releaseNotesFile).
		WithNewFile(versionPath, version+"\n").
		With(func(d *dagger.Directory) *dagger.Directory {
			if chartFile != nil {
				d = d.WithFile(path.Join(chartPath, "Chart.yaml"), chartFile)
			}
			return d
		})
	return after.Changes(src), nil
}

// regex to confirm valid semver
var semverRegex = regexp.MustCompile(`(\d+\.\d+\.\d+(?:-[0-9A-Za-z-.]+)?)`)

// Generate the next semantic version from conventional commit messages (see cliff.toml).
// The returned version is of the form MAJOR.MINOR.PATCH.
func (r *Release) Version(ctx context.Context,
	//git-cliff cliff.toml path to use. Defaults to root of gitref "./"
	// +optional
	config string,
) (string, error) {

	bumpedTag, err := dag.GitCliff(r.GitRef).
		BumpedVersion(ctx, dagger.GitCliffBumpedVersionOpts{Config: config})

	semver := semverRegex.FindStringSubmatch(bumpedTag)

	if len(semver) < 2 {
		return "", fmt.Errorf("valid semver not found in: %s", bumpedTag)
	}
	return semver[1], err
}

// Generate the change log from conventional commit messages.
//
// changelog is a default changelog generated using the git-cliff module. Please use the act3-ai/dagger/git-cliff module directly for custom changelogs.
func (r *Release) changelog(
	ctx context.Context,
	//version to generate changelog for
	version string,
	// Changelog file path, relative to source directory
	// +default="CHANGELOG.md"
	changelog string,
	//git-cliff cliff.toml path to use. Defaults to root of gitref "./"
	config string,
	//provide a github token to git-cliff.
	//This is needed due to github api rate limit
	token *dagger.Secret,
) *dagger.File {
	version = strings.TrimPrefix(version, "v")

	// generate and prepend to changelog
	return dag.GitCliff(r.GitRef).
		WithTag(version).
		WithStrip("footer").
		WithUnreleased().
		With(func(gc *dagger.GitCliff) *dagger.GitCliff {
			//use alternate git-cliff config if provided
			if config != "" {
				gc = gc.WithConfig(dagger.GitCliffWithConfigOpts{Config: config})
			}
			// use token if provided
			if token != nil {
				gc = gc.WithSecretVariable("GITHUB_TOKEN", token)
			}

			// check if changelog file exists, if not create it
			exists, err := r.GitRef.Tree().Exists(ctx, changelog)
			if err != nil {
				panic(fmt.Errorf("failed to check if %s exists: %w", changelog, err))
			}

			if !exists {
				return gc.WithOutput(changelog)
			}

			// if file exists, prepend instead
			return gc.WithPrepend(changelog)
		}).
		Run().
		File(changelog)
}

// Generate release notes.
//
// notes are default release notes generated using the git-cliff module. Please use the act3-ai/dagger/git-cliff module directly for custom release notes.
func (r *Release) notes(ctx context.Context,
	version string,
	// Custom release notes file name. Default: v<version>.md
	name string,
	// Additional information to include in release notes. Injected after header and before commit
	extraNotes string,
	//git-cliff cliff.toml path to use. Defaults to root of gitref "./"
	config string,
	//provide a github token to git-cliff.
	//This is needed due to github api rate limit
	token *dagger.Secret,
) (*dagger.File, error) {
	version = strings.TrimPrefix(version, "v")

	// generate and export release notes
	notes, err := dag.GitCliff(r.GitRef).
		WithTag(version).
		WithUnreleased().
		WithStrip("all").
		With(func(gc *dagger.GitCliff) *dagger.GitCliff {
			// use alternate git-cliff config if provided
			if config != "" {
				gc = gc.WithConfig(dagger.GitCliffWithConfigOpts{Config: config})
			}
			// use token if provided
			if token != nil {
				gc = gc.WithSecretVariable("GITHUB_TOKEN", token)
			}
			return gc
		}).
		Run().
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
	// path to the chart
	chartPath string,
) *dagger.File {
	version = strings.TrimPrefix(version, "v")
	file := path.Join(chartPath, "Chart.yaml")
	return dag.Wolfi().
		Container(dagger.WolfiContainerOpts{
			Packages: []string{"yq"},
		}).
		WithMountedDirectory("/src", r.GitRef.Tree()).
		WithWorkdir("/src").
		WithEnvVariable("version", version).
		WithExec([]string{"yq", "e",
			"(.version = env(version)) | (.appVersion = \"v\"+env(version))",
			"-i", file}).
		File(file)
}
