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

	// Prepend path prefix if given
	changelogPath := filepath.Join(pathPrefix, "CHANGELOG.md")
	versionPath := filepath.Join(pathPrefix, "VERSION")
	notesPath := filepath.Join(pathPrefix, "releases", fmt.Sprintf("v%s.md", version))

	// generate changelog if one is not found, else prepend to an existing one
	chlogOpts := dagger.GitCliffDevOpts{
		Tag:    version,
		Config: config,
		Strip:  "footer",
	}

	if r.changelogCheck(ctx, changelogPath) {
		chlogOpts.Prepend = changelogPath
	} else {
		chlogOpts.OutputFile = changelogPath
	}

	changelogFile := dag.GitCliffDev(r.GitRef, chlogOpts).Run().File(changelogPath)

	// generate release notes with optional extraNotes if provided
	releaseNotesOpts := dagger.GitCliffDevOpts{
		Tag:    version,
		Config: config,
		Strip:  "all",
	}

	releaseNotes, err := dag.GitCliffDev(r.GitRef, releaseNotesOpts).Run().Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("generating release notes: %w", err)
	}

	if extraNotes != "" {
		releaseNotes = strings.Replace(releaseNotes, "###", extraNotes+"###", 1)
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
		WithNewFile(notesPath, releaseNotes).
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

// check if changelog exists
func (r *Release) changelogCheck(
	ctx context.Context,
	// path to the changelog
	changelogPath string,
) bool {

	exists, err := r.GitRef.Tree().Exists(ctx, changelogPath)
	if err != nil {
		panic(fmt.Errorf("failed to check if %s exists: %w", changelogPath, err))
	}
	return exists
}
